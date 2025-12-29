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
	removeNullValues(result)

	return result
}

// removeNullValues recursively removes null values from a map.
func removeNullValues(m map[string]interface{}) {
	for key, value := range m {
		if value == nil {
			delete(m, key)
		} else if nested, ok := value.(map[string]interface{}); ok {
			removeNullValues(nested)
			if len(nested) == 0 {
				delete(m, key)
			}
		}
	}
}

// deepMergeMap merges src into dst, with src taking precedence.
func deepMergeMap(dst, src map[string]interface{}) map[string]interface{} {
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
				result[k] = deepMergeMap(dstMap, srcMap)
			} else {
				result[k] = srcMap
			}
		} else {
			result[k] = v
		}
	}

	return result
}

// Wrapper method on reconciler to maintain existing API
func (r *ImmichReconciler) deepMergeMap(dst, src map[string]interface{}) map[string]interface{} {
	return deepMergeMap(dst, src)
}
