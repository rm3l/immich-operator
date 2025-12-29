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
	"fmt"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

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
		postgres := immich.Spec.Postgres
		if postgres == nil || postgres.Host == "" {
			configErrors = append(configErrors, "spec.postgres.host is required when spec.postgres.enabled=false")
		}
		if postgres == nil || (postgres.PasswordSecretRef == nil && postgres.URLSecretRef == nil) {
			configErrors = append(configErrors, "spec.postgres.password or spec.postgres.passwordSecretRef is required when spec.postgres.enabled=false")
		}
	}
	// Note: When postgres.enabled=true and no password is provided, the operator auto-generates credentials

	// Validate external Valkey config when built-in is disabled
	if !immich.IsValkeyEnabled() {
		valkey := immich.Spec.Valkey
		if valkey == nil || valkey.Host == "" {
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
