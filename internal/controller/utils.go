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

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// FieldManager is the field manager name used for server-side apply
const FieldManager = "immich-operator"

// mergeMaps merges two string maps, with override taking precedence
func mergeMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// Wrapper method on reconciler to maintain existing API
func (r *ImmichReconciler) mergeMaps(base, override map[string]string) map[string]string {
	return mergeMaps(base, override)
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

// apply uses server-side apply to create or update a resource.
// The object must have its GVK set (TypeMeta populated).
// Server-side apply provides:
// - Better conflict resolution with field ownership tracking
// - No need to read-before-write (eliminates race conditions)
// - Declarative updates where only specified fields are managed
func (r *ImmichReconciler) apply(ctx context.Context, obj client.Object) error {
	log := logf.FromContext(ctx)

	err := r.Patch(ctx, obj, client.Apply, client.FieldOwner(FieldManager), client.ForceOwnership)
	if err != nil {
		return err
	}

	log.V(1).Info("Applied resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
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
