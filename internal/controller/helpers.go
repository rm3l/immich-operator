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

// reconcileImmichConfig creates or updates the Immich configuration ConfigMap or Secret
func (r *ImmichReconciler) reconcileImmichConfig(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)

	if immich.Spec.Immich.Configuration == nil {
		log.V(1).Info("No Immich configuration specified, skipping ConfigMap/Secret creation")
		return nil
	}

	configName := fmt.Sprintf("%s-immich-config", immich.Name)

	// Convert configuration to YAML
	configData, err := yaml.Marshal(immich.Spec.Immich.Configuration)
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

// reconcileLibraryPVC creates the PVC for the photo library if needed
func (r *ImmichReconciler) reconcileLibraryPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Library PVC")

	name := immich.GetLibraryPVCName()
	labels := r.getLabels(immich, "library")

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(immich, pvc, r.Scheme); err != nil {
		return err
	}

	// Check if PVC already exists - we can't update it
	existing := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, existing)
	if err == nil {
		// PVC exists, don't update
		log.V(1).Info("Library PVC already exists", "name", name)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// Create new PVC
	var storageClassName *string
	if immich.Spec.Immich.Persistence.Library.StorageClass != "" {
		storageClassName = &immich.Spec.Immich.Persistence.Library.StorageClass
	}

	size := immich.GetLibrarySize()
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes:      immich.GetLibraryAccessModes(),
		StorageClassName: storageClassName,
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: size,
			},
		},
	}

	log.Info("Creating Library PVC", "name", name, "size", size.String())
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

	// Validate external ML config when built-in is disabled
	if !immich.IsMachineLearningEnabled() {
		if immich.Spec.MachineLearning.URL == "" {
			configErrors = append(configErrors, "spec.machineLearning.url is required when spec.machineLearning.enabled=false")
		}
	}

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
