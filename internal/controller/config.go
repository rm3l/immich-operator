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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

// reconcileImmichConfig creates or updates the Immich configuration ConfigMap or Secret using server-side apply.
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
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      configName,
				Namespace: immich.Namespace,
				Labels:    labels,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion:         immich.APIVersion,
						Kind:               immich.Kind,
						Name:               immich.Name,
						UID:                immich.UID,
						Controller:         ptr.To(true),
						BlockOwnerDeletion: ptr.To(true),
					},
				},
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"immich-config.yaml": string(configData),
			},
		}

		return r.apply(ctx, secret)
	}

	// Default to ConfigMap
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: immich.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         immich.APIVersion,
					Kind:               immich.Kind,
					Name:               immich.Name,
					UID:                immich.UID,
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				},
			},
		},
		Data: map[string]string{
			"immich-config.yaml": string(configData),
		},
	}

	return r.apply(ctx, configMap)
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
	removeNullValues(result)

	return result
}
