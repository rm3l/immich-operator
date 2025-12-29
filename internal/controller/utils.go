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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

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
