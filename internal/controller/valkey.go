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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// reconcileValkey creates or updates the Valkey (Redis) deployment and service
func (r *ImmichReconciler) reconcileValkey(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Valkey")

	// Create Valkey Deployment
	if err := r.reconcileValkeyDeployment(ctx, immich); err != nil {
		return err
	}

	// Create Valkey Service
	if err := r.reconcileValkeyService(ctx, immich); err != nil {
		return err
	}

	// Create Valkey PVC if persistence is enabled
	if immich.Spec.Valkey.Persistence.Enabled != nil && *immich.Spec.Valkey.Persistence.Enabled {
		if err := r.reconcileValkeyPVC(ctx, immich); err != nil {
			return err
		}
	}

	return nil
}

func (r *ImmichReconciler) reconcileValkeyDeployment(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-valkey", immich.Name)
	labels := r.getLabels(immich, "valkey")
	selectorLabels := r.getSelectorLabels(immich, "valkey")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(immich, deployment, r.Scheme); err != nil {
		return err
	}

	return r.createOrUpdate(ctx, deployment, func() error {
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.mergeMaps(labels, immich.Spec.Valkey.PodLabels),
					Annotations: immich.Spec.Valkey.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					SecurityContext:  immich.Spec.Valkey.PodSecurityContext,
					ImagePullSecrets: immich.Spec.ImagePullSecrets,
					NodeSelector:     immich.Spec.Valkey.NodeSelector,
					Tolerations:      immich.Spec.Valkey.Tolerations,
					Affinity:         immich.Spec.Valkey.Affinity,
					Containers: []corev1.Container{
						{
							Name:            "valkey",
							Image:           immich.GetValkeyImage(),
							ImagePullPolicy: immich.Spec.Valkey.ImagePullPolicy,
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: 6379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources:       immich.Spec.Valkey.Resources,
							SecurityContext: immich.Spec.Valkey.SecurityContext,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"sh", "-c", "valkey-cli ping | grep PONG"},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"sh", "-c", "valkey-cli ping | grep PONG"},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							VolumeMounts: r.getValkeyVolumeMounts(immich),
						},
					},
					Volumes: r.getValkeyVolumes(immich),
				},
			},
		}
		return nil
	})
}

func (r *ImmichReconciler) getValkeyVolumeMounts(immich *mediav1alpha1.Immich) []corev1.VolumeMount {
	if immich.Spec.Valkey.Persistence.Enabled != nil && *immich.Spec.Valkey.Persistence.Enabled {
		return []corev1.VolumeMount{
			{
				Name:      "data",
				MountPath: "/data",
			},
		}
	}
	return nil
}

func (r *ImmichReconciler) getValkeyVolumes(immich *mediav1alpha1.Immich) []corev1.Volume {
	if immich.Spec.Valkey.Persistence.Enabled != nil && *immich.Spec.Valkey.Persistence.Enabled {
		pvcName := fmt.Sprintf("%s-valkey-data", immich.Name)
		if immich.Spec.Valkey.Persistence.ExistingClaim != "" {
			pvcName = immich.Spec.Valkey.Persistence.ExistingClaim
		}
		return []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvcName,
					},
				},
			},
		}
	}
	return []corev1.Volume{
		{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func (r *ImmichReconciler) reconcileValkeyService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-valkey", immich.Name)
	labels := r.getLabels(immich, "valkey")
	selectorLabels := r.getSelectorLabels(immich, "valkey")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(immich, service, r.Scheme); err != nil {
		return err
	}

	return r.createOrUpdate(ctx, service, func() error {
		service.Spec = corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selectorLabels,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       6379,
					TargetPort: intstr.FromString("redis"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}
		return nil
	})
}

func (r *ImmichReconciler) reconcileValkeyPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	if immich.Spec.Valkey.Persistence.ExistingClaim != "" {
		return nil // Using existing PVC
	}

	name := fmt.Sprintf("%s-valkey-data", immich.Name)
	labels := r.getLabels(immich, "valkey")

	size := immich.Spec.Valkey.Persistence.Size
	if size.IsZero() {
		size = resource.MustParse("1Gi")
	}

	accessModes := immich.Spec.Valkey.Persistence.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(immich, pvc, r.Scheme); err != nil {
		return err
	}

	// Check if PVC already exists - we can't update it
	existing := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, existing)
	if err == nil {
		// PVC exists, don't update
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// Create new PVC
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes:      accessModes,
		StorageClassName: immich.Spec.Valkey.Persistence.StorageClass,
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: size,
			},
		},
	}

	return r.Create(ctx, pvc)
}
