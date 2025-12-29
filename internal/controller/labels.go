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
	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// Standard Kubernetes labels
const (
	labelApp       = "app.kubernetes.io/name"
	labelInstance  = "app.kubernetes.io/instance"
	labelComponent = "app.kubernetes.io/component"
	labelManagedBy = "app.kubernetes.io/managed-by"
	labelPartOf    = "app.kubernetes.io/part-of"
)

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
