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

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

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
