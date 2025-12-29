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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// reconcileLibraryPVC creates the PVC for the photo library if needed.
// Note: Library PVCs do NOT have an owner reference to allow data persistence
// across Immich CR deletions and recreations.
func (r *ImmichReconciler) reconcileLibraryPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Library PVC")

	name := immich.GetLibraryPVCName()
	labels := r.getLabels(immich, "library")

	// Check if PVC already exists - reuse it if so
	existing := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, existing)
	if err == nil {
		// PVC exists, reuse it (don't update - PVCs are mostly immutable)
		log.V(1).Info("Library PVC already exists, reusing", "name", name)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// Create new PVC (without owner reference for data safety)
	storageClassName := immich.GetLibraryStorageClass()

	size := immich.GetLibrarySize()
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      immich.GetLibraryAccessModes(),
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: size,
				},
			},
		},
	}

	// Note: We intentionally do NOT set owner reference here.
	// This ensures the PVC persists when the Immich CR is deleted,
	// protecting user data and allowing reuse on CR recreation.

	log.Info("Creating Library PVC (no owner reference for data safety)", "name", name, "size", size.String())
	return r.Create(ctx, pvc)
}
