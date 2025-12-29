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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

func TestGetLabels(t *testing.T) {
	r := &ImmichReconciler{}
	immich := &mediav1alpha1.Immich{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-immich",
			Namespace: "default",
		},
	}

	tests := []struct {
		name      string
		component string
		expected  map[string]string
	}{
		{
			name:      "server component",
			component: "server",
			expected: map[string]string{
				"app.kubernetes.io/name":       "immich",
				"app.kubernetes.io/instance":   "test-immich",
				"app.kubernetes.io/component":  "server",
				"app.kubernetes.io/managed-by": "immich-operator",
				"app.kubernetes.io/part-of":    "immich",
			},
		},
		{
			name:      "machine-learning component",
			component: "machine-learning",
			expected: map[string]string{
				"app.kubernetes.io/name":       "immich",
				"app.kubernetes.io/instance":   "test-immich",
				"app.kubernetes.io/component":  "machine-learning",
				"app.kubernetes.io/managed-by": "immich-operator",
				"app.kubernetes.io/part-of":    "immich",
			},
		},
		{
			name:      "postgres component",
			component: "postgres",
			expected: map[string]string{
				"app.kubernetes.io/name":       "immich",
				"app.kubernetes.io/instance":   "test-immich",
				"app.kubernetes.io/component":  "postgres",
				"app.kubernetes.io/managed-by": "immich-operator",
				"app.kubernetes.io/part-of":    "immich",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.getLabels(immich, tt.component)
			if !stringMapsEqual(result, tt.expected) {
				t.Errorf("getLabels() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetSelectorLabels(t *testing.T) {
	r := &ImmichReconciler{}
	immich := &mediav1alpha1.Immich{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-immich",
			Namespace: "default",
		},
	}

	tests := []struct {
		name      string
		component string
		expected  map[string]string
	}{
		{
			name:      "server component",
			component: "server",
			expected: map[string]string{
				"app.kubernetes.io/name":      "immich",
				"app.kubernetes.io/instance":  "test-immich",
				"app.kubernetes.io/component": "server",
			},
		},
		{
			name:      "valkey component",
			component: "valkey",
			expected: map[string]string{
				"app.kubernetes.io/name":      "immich",
				"app.kubernetes.io/instance":  "test-immich",
				"app.kubernetes.io/component": "valkey",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.getSelectorLabels(immich, tt.component)
			if !stringMapsEqual(result, tt.expected) {
				t.Errorf("getSelectorLabels() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// stringMapsEqual compares two string maps
func stringMapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			return false
		}
	}
	return true
}
