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
	"os"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

func TestValidateImages(t *testing.T) {
	// Set up environment variables for image defaults
	os.Setenv(mediav1alpha1.EnvRelatedImageImmich, "ghcr.io/immich-app/immich-server:latest")
	os.Setenv(mediav1alpha1.EnvRelatedImageMachineLearning, "ghcr.io/immich-app/immich-machine-learning:latest")
	os.Setenv(mediav1alpha1.EnvRelatedImageValkey, "docker.io/valkey/valkey:9-alpine")
	os.Setenv(mediav1alpha1.EnvRelatedImagePostgres, "docker.io/tensorchord/pgvecto-rs:pg17-v0.4.0")
	defer func() {
		os.Unsetenv(mediav1alpha1.EnvRelatedImageImmich)
		os.Unsetenv(mediav1alpha1.EnvRelatedImageMachineLearning)
		os.Unsetenv(mediav1alpha1.EnvRelatedImageValkey)
		os.Unsetenv(mediav1alpha1.EnvRelatedImagePostgres)
	}()

	r := &ImmichReconciler{}

	tests := []struct {
		name        string
		immich      *mediav1alpha1.Immich
		expectError bool
		errorSubstr string
	}{
		{
			name: "all defaults with env vars",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{},
			},
			expectError: false,
		},
		{
			name: "external postgres without host",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					Postgres: mediav1alpha1.PostgresSpec{
						Enabled: boolPtr(false),
					},
				},
			},
			expectError: true,
			errorSubstr: "spec.postgres.host is required",
		},
		{
			name: "external postgres without password",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					Postgres: mediav1alpha1.PostgresSpec{
						Enabled: boolPtr(false),
						Host:    "external-postgres.example.com",
					},
				},
			},
			expectError: true,
			errorSubstr: "passwordSecretRef is required",
		},
		{
			name: "external postgres with proper config",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					Postgres: mediav1alpha1.PostgresSpec{
						Enabled: boolPtr(false),
						Host:    "external-postgres.example.com",
						PasswordSecretRef: &mediav1alpha1.SecretKeySelector{
							Name: "postgres-secret",
							Key:  "password",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "external valkey without host",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					Valkey: mediav1alpha1.ValkeySpec{
						Enabled: boolPtr(false),
					},
				},
			},
			expectError: true,
			errorSubstr: "spec.valkey.host is required",
		},
		{
			name: "external valkey with host",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					Valkey: mediav1alpha1.ValkeySpec{
						Enabled: boolPtr(false),
						Host:    "external-redis.example.com",
					},
				},
			},
			expectError: false,
		},
		{
			name: "ML disabled without URL is valid",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					MachineLearning: mediav1alpha1.MachineLearningSpec{
						Enabled: boolPtr(false),
					},
				},
			},
			expectError: false,
		},
		{
			name: "ML disabled with external URL is valid",
			immich: &mediav1alpha1.Immich{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-immich",
					Namespace: "default",
				},
				Spec: mediav1alpha1.ImmichSpec{
					MachineLearning: mediav1alpha1.MachineLearningSpec{
						Enabled: boolPtr(false),
						URL:     "http://external-ml.example.com:3003",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateImages(tt.immich)
			if tt.expectError {
				if err == nil {
					t.Errorf("validateImages() expected error, got nil")
				} else if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("validateImages() error = %v, expected to contain %q", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("validateImages() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateImages_MissingEnvVars(t *testing.T) {
	// Clear all RELATED_IMAGE env vars
	os.Unsetenv(mediav1alpha1.EnvRelatedImageImmich)
	os.Unsetenv(mediav1alpha1.EnvRelatedImageMachineLearning)
	os.Unsetenv(mediav1alpha1.EnvRelatedImageValkey)
	os.Unsetenv(mediav1alpha1.EnvRelatedImagePostgres)

	r := &ImmichReconciler{}

	immich := &mediav1alpha1.Immich{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-immich",
			Namespace: "default",
		},
		Spec: mediav1alpha1.ImmichSpec{},
	}

	err := r.validateImages(immich)
	if err == nil {
		t.Error("validateImages() expected error when env vars are not set")
	}
	if !strings.Contains(err.Error(), "missing required images") {
		t.Errorf("validateImages() error = %v, expected to mention missing images", err)
	}
}

func TestValidateImages_WithSpecImages(t *testing.T) {
	// Clear all RELATED_IMAGE env vars
	os.Unsetenv(mediav1alpha1.EnvRelatedImageImmich)
	os.Unsetenv(mediav1alpha1.EnvRelatedImageMachineLearning)
	os.Unsetenv(mediav1alpha1.EnvRelatedImageValkey)
	os.Unsetenv(mediav1alpha1.EnvRelatedImagePostgres)

	r := &ImmichReconciler{}

	immich := &mediav1alpha1.Immich{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-immich",
			Namespace: "default",
		},
		Spec: mediav1alpha1.ImmichSpec{
			Server: mediav1alpha1.ServerSpec{
				Image: "ghcr.io/immich-app/immich-server:v1.0.0",
			},
			MachineLearning: mediav1alpha1.MachineLearningSpec{
				Image: "ghcr.io/immich-app/immich-machine-learning:v1.0.0",
			},
			Valkey: mediav1alpha1.ValkeySpec{
				Image: "docker.io/valkey/valkey:9-alpine",
			},
			Postgres: mediav1alpha1.PostgresSpec{
				Image: "docker.io/tensorchord/pgvecto-rs:pg17-v0.4.0",
			},
		},
	}

	err := r.validateImages(immich)
	if err != nil {
		t.Errorf("validateImages() unexpected error = %v", err)
	}
}

// boolPtr returns a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
