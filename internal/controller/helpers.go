/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"crypto/rand"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

// reconcileImmichConfig creates or updates the Immich configuration ConfigMap or Secret.
// It builds a base configuration from CR state and merges it with user-provided configuration.
func (r *ImmichReconciler) reconcileImmichConfig(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Immich configuration")

	configName := fmt.Sprintf("%s-immich-config", immich.Name)

	// Build effective configuration by merging base config with user config
	effectiveConfig := r.buildEffectiveConfigMap(immich)

	// Convert configuration to YAML
	configData, err := yaml.Marshal(effectiveConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal immich configuration: %w", err)
	}

	labels := r.getLabels(immich, "config")

	if immich.Spec.Immich.ConfigurationKind == "Secret" {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configName,
				Namespace: immich.Namespace,
				Labels:    labels,
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"immich-config.yaml": string(configData),
			},
		}

		if err := controllerutil.SetControllerReference(immich, secret, r.Scheme); err != nil {
			return err
		}

		return r.createOrUpdate(ctx, secret, func() error {
			secret.StringData = map[string]string{
				"immich-config.yaml": string(configData),
			}
			return nil
		})
	}

	// Default to ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"immich-config.yaml": string(configData),
		},
	}

	if err := controllerutil.SetControllerReference(immich, configMap, r.Scheme); err != nil {
		return err
	}

	return r.createOrUpdate(ctx, configMap, func() error {
		configMap.Data = map[string]string{
			"immich-config.yaml": string(configData),
		}
		return nil
	})
}

// buildEffectiveConfigMap builds the effective Immich configuration as a map.
// This avoids issues with nil struct fields being marshaled as null.
// User configuration takes precedence over operator-derived settings.
func (r *ImmichReconciler) buildEffectiveConfigMap(immich *mediav1alpha1.Immich) map[string]interface{} {
	config := make(map[string]interface{})

	// First, apply operator-derived ML settings
	r.applyMLConfigMap(immich, config)

	// Then, merge user-provided configuration (user takes precedence)
	if immich.Spec.Immich.Configuration != nil {
		userConfig := r.configSpecToMap(immich.Spec.Immich.Configuration)
		config = r.deepMergeMap(config, userConfig)
	}

	return config
}

// applyMLConfigMap applies machine learning configuration based on CR state.
// Follows the Immich config structure: https://docs.immich.app/install/config-file/
func (r *ImmichReconciler) applyMLConfigMap(immich *mediav1alpha1.Immich, config map[string]interface{}) {
	// Get the ML URL (built-in service URL, external URL, or empty if disabled)
	mlURL := immich.GetMachineLearningURL()

	// Determine if ML should be enabled
	// ML is enabled if: built-in is enabled OR external URL is provided
	mlEnabled := immich.IsMachineLearningEnabled() || immich.Spec.MachineLearning.URL != ""

	// Build ML config map with only non-empty values
	// Note: Immich uses "urls" (array) not "url" (string)
	mlConfig := map[string]interface{}{
		"enabled": mlEnabled,
	}
	if mlURL != "" {
		mlConfig["urls"] = []string{mlURL}
	}

	config["machineLearning"] = mlConfig
}

// configSpecToMap converts a ConfigurationSpec to a map, excluding nil fields.
func (r *ImmichReconciler) configSpecToMap(spec *mediav1alpha1.ConfigurationSpec) map[string]interface{} {
	// Marshal to YAML then unmarshal to map to get a clean representation
	// This automatically handles omitempty and excludes nil fields
	data, err := yaml.Marshal(spec)
	if err != nil {
		return make(map[string]interface{})
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return make(map[string]interface{})
	}

	// Remove any null values that might have slipped through
	r.removeNullValues(result)

	return result
}

// removeNullValues recursively removes null values from a map
func (r *ImmichReconciler) removeNullValues(m map[string]interface{}) {
	for key, value := range m {
		if value == nil {
			delete(m, key)
		} else if nested, ok := value.(map[string]interface{}); ok {
			r.removeNullValues(nested)
			if len(nested) == 0 {
				delete(m, key)
			}
		}
	}
}

// deepMergeMap merges src into dst, with src taking precedence
func (r *ImmichReconciler) deepMergeMap(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy dst
	for k, v := range dst {
		result[k] = v
	}

	// Merge src (overrides dst)
	for k, v := range src {
		if v == nil {
			continue
		}
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := result[k].(map[string]interface{}); ok {
				result[k] = r.deepMergeMap(dstMap, srcMap)
			} else {
				result[k] = srcMap
			}
		} else {
			result[k] = v
		}
	}

	return result
}

// reconcileLibraryPVC creates the PVC for the photo library if needed.
// Note: Library PVCs do NOT have an owner reference to allow data persistence
// across Immich CR deletions and recreations.
func (r *ImmichReconciler) reconcileLibraryPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Library PVC")

	name := immich.GetLibraryPVCName()
	labels := r.getLabels(immich, "library")

	// Check if PVC already exists - reuse it if so
	existing := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, existing)
	if err == nil {
		// PVC exists, reuse it (don't update - PVCs are mostly immutable)
		log.V(1).Info("Library PVC already exists, reusing", "name", name)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// Create new PVC (without owner reference for data safety)
	var storageClassName *string
	if immich.Spec.Immich.Persistence.Library.StorageClass != "" {
		storageClassName = &immich.Spec.Immich.Persistence.Library.StorageClass
	}

	size := immich.GetLibrarySize()
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      immich.GetLibraryAccessModes(),
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: size,
				},
			},
		},
	}

	// Note: We intentionally do NOT set owner reference here.
	// This ensures the PVC persists when the Immich CR is deleted,
	// protecting user data and allowing reuse on CR recreation.

	log.Info("Creating Library PVC (no owner reference for data safety)", "name", name, "size", size.String())
	return r.Create(ctx, pvc)
}

// updateStatus updates the status of the Immich resource
func (r *ImmichReconciler) updateStatus(ctx context.Context, immich *mediav1alpha1.Immich) error {
	// Check Server status
	if immich.IsServerEnabled() {
		deployment := &appsv1.Deployment{}
		name := fmt.Sprintf("%s-server", immich.Name)
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, deployment); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			immich.Status.ServerReady = false
		} else {
			immich.Status.ServerReady = deployment.Status.ReadyReplicas > 0 &&
				deployment.Status.ReadyReplicas == deployment.Status.Replicas
		}
	} else {
		immich.Status.ServerReady = true
	}

	// Check ML status
	if immich.IsMachineLearningEnabled() {
		deployment := &appsv1.Deployment{}
		name := fmt.Sprintf("%s-machine-learning", immich.Name)
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, deployment); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			immich.Status.MachineLearningReady = false
		} else {
			immich.Status.MachineLearningReady = deployment.Status.ReadyReplicas > 0 &&
				deployment.Status.ReadyReplicas == deployment.Status.Replicas
		}
	} else {
		immich.Status.MachineLearningReady = true
	}

	// Check Valkey status
	if immich.IsValkeyEnabled() {
		deployment := &appsv1.Deployment{}
		name := fmt.Sprintf("%s-valkey", immich.Name)
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, deployment); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			immich.Status.ValkeyReady = false
		} else {
			immich.Status.ValkeyReady = deployment.Status.ReadyReplicas > 0 &&
				deployment.Status.ReadyReplicas == deployment.Status.Replicas
		}
	} else {
		immich.Status.ValkeyReady = true
	}

	// Check PostgreSQL status
	if immich.IsPostgresEnabled() {
		sts := &appsv1.StatefulSet{}
		name := fmt.Sprintf("%s-postgres", immich.Name)
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, sts); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
			immich.Status.PostgresReady = false
		} else {
			immich.Status.PostgresReady = sts.Status.ReadyReplicas > 0 &&
				sts.Status.ReadyReplicas == sts.Status.Replicas
		}
	} else {
		immich.Status.PostgresReady = true
	}

	// Overall ready status
	immich.Status.Ready = immich.Status.ServerReady &&
		immich.Status.MachineLearningReady &&
		immich.Status.ValkeyReady &&
		immich.Status.PostgresReady

	return nil
}

// validateImages checks that all required images are configured
func (r *ImmichReconciler) validateImages(immich *mediav1alpha1.Immich) error {
	var missingImages []string
	var configErrors []string

	if immich.IsServerEnabled() && immich.GetServerImage() == "" {
		missingImages = append(missingImages, fmt.Sprintf("server (set spec.server.image or %s env var)", mediav1alpha1.EnvRelatedImageImmich))
	}

	if immich.IsMachineLearningEnabled() && immich.GetMachineLearningImage() == "" {
		missingImages = append(missingImages, fmt.Sprintf("machine-learning (set spec.machineLearning.image or %s env var)", mediav1alpha1.EnvRelatedImageMachineLearning))
	}

	if immich.IsValkeyEnabled() && immich.GetValkeyImage() == "" {
		missingImages = append(missingImages, fmt.Sprintf("valkey (set spec.valkey.image or %s env var)", mediav1alpha1.EnvRelatedImageValkey))
	}

	if immich.IsPostgresEnabled() && immich.GetPostgresImage() == "" {
		missingImages = append(missingImages, fmt.Sprintf("postgres (set spec.postgres.image or %s env var)", mediav1alpha1.EnvRelatedImagePostgres))
	}

	// Validate external PostgreSQL config when built-in is disabled
	if !immich.IsPostgresEnabled() {
		if immich.Spec.Postgres.Host == "" {
			configErrors = append(configErrors, "spec.postgres.host is required when spec.postgres.enabled=false")
		}
		if immich.Spec.Postgres.PasswordSecretRef == nil && immich.Spec.Postgres.URLSecretRef == nil {
			configErrors = append(configErrors, "spec.postgres.password or spec.postgres.passwordSecretRef is required when spec.postgres.enabled=false")
		}
	}
	// Note: When postgres.enabled=true and no password is provided, the operator auto-generates credentials

	// Validate external Valkey config when built-in is disabled
	if !immich.IsValkeyEnabled() {
		if immich.Spec.Valkey.Host == "" {
			configErrors = append(configErrors, "spec.valkey.host is required when spec.valkey.enabled=false")
		}
	}

	// Note: Machine Learning is optional - it can be disabled completely without providing an external URL.
	// When disabled without an external URL, Immich will run without ML features (smart search, face detection, etc.).

	if len(missingImages) > 0 {
		return fmt.Errorf("missing required images: %v", missingImages)
	}

	if len(configErrors) > 0 {
		return fmt.Errorf("configuration errors: %v", configErrors)
	}

	return nil
}

// getLabels returns the standard labels for Immich components
func (r *ImmichReconciler) getLabels(immich *mediav1alpha1.Immich, component string) map[string]string {
	return map[string]string{
		labelApp:       "immich",
		labelInstance:  immich.Name,
		labelComponent: component,
		labelManagedBy: "immich-operator",
		labelPartOf:    "immich",
	}
}

// getSelectorLabels returns the selector labels for Immich components
func (r *ImmichReconciler) getSelectorLabels(immich *mediav1alpha1.Immich, component string) map[string]string {
	return map[string]string{
		labelApp:       "immich",
		labelInstance:  immich.Name,
		labelComponent: component,
	}
}

// mergeMaps merges two string maps, with override taking precedence
func (r *ImmichReconciler) mergeMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// createOrUpdate wraps controllerutil.CreateOrUpdate with logging
func (r *ImmichReconciler) createOrUpdate(ctx context.Context, obj client.Object, mutate func() error) error {
	log := logf.FromContext(ctx)

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, mutate)
	if err != nil {
		return err
	}

	switch result {
	case controllerutil.OperationResultCreated:
		log.Info("Created resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
	case controllerutil.OperationResultUpdated:
		log.V(1).Info("Updated resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
	}

	return nil
}

// generateRandomPassword generates a cryptographically secure random password
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
