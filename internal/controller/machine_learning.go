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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// reconcileMachineLearning creates or updates the Machine Learning deployment and service
func (r *ImmichReconciler) reconcileMachineLearning(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Machine Learning")

	mlSpec := ptr.Deref(immich.Spec.MachineLearning, mediav1alpha1.MachineLearningSpec{})
	persistence := ptr.Deref(mlSpec.Persistence, mediav1alpha1.MachineLearningPersistenceSpec{})

	// Create ML PVC first if persistence is enabled (must exist before deployment)
	// Default is enabled (nil or true)
	persistenceEnabled := persistence.Enabled == nil || *persistence.Enabled
	if persistenceEnabled {
		if err := r.reconcileMLPVC(ctx, immich); err != nil {
			return err
		}
	}

	// Create ML Deployment
	if err := r.reconcileMLDeployment(ctx, immich); err != nil {
		return err
	}

	// Create ML Service
	if err := r.reconcileMLService(ctx, immich); err != nil {
		return err
	}

	return nil
}

// reconcileMLDeployment creates or updates the ML Deployment using server-side apply
func (r *ImmichReconciler) reconcileMLDeployment(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-machine-learning", immich.Name)
	labels := r.getLabels(immich, "machine-learning")
	selectorLabels := r.getSelectorLabels(immich, "machine-learning")

	mlSpec := ptr.Deref(immich.Spec.MachineLearning, mediav1alpha1.MachineLearningSpec{})
	replicas := ptr.Deref(mlSpec.Replicas, 1)

	env := []corev1.EnvVar{
		{Name: "TRANSFORMERS_CACHE", Value: "/cache"},
		{Name: "HF_XET_CACHE", Value: "/cache/huggingface-xet"},
		{Name: "MPLCONFIGDIR", Value: "/cache/matplotlib-config"},
	}
	env = append(env, mlSpec.Env...)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         immich.APIVersion,
					Kind:               immich.Kind,
					Name:               immich.Name,
					UID:                immich.UID,
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.mergeMaps(labels, mlSpec.PodLabels),
					Annotations: mlSpec.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					SecurityContext:  mlSpec.PodSecurityContext,
					ImagePullSecrets: immich.Spec.ImagePullSecrets,
					NodeSelector:     mlSpec.NodeSelector,
					Tolerations:      mlSpec.Tolerations,
					Affinity:         mlSpec.Affinity,
					Containers: []corev1.Container{
						{
							Name:            "machine-learning",
							Image:           immich.GetMachineLearningImage(),
							ImagePullPolicy: mlSpec.ImagePullPolicy,
							Env:             env,
							EnvFrom:         mlSpec.EnvFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 3003,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources:       mlSpec.Resources,
							SecurityContext: mlSpec.SecurityContext,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ping",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ping",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								FailureThreshold:    3,
							},
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ping",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								FailureThreshold:    60,
							},
							VolumeMounts: r.getMLVolumeMounts(immich),
						},
					},
					Volumes: r.getMLVolumes(immich),
				},
			},
		},
	}

	return r.apply(ctx, deployment)
}

func (r *ImmichReconciler) getMLVolumeMounts(_ *mediav1alpha1.Immich) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "cache",
			MountPath: "/cache",
		},
	}
}

func (r *ImmichReconciler) getMLVolumes(immich *mediav1alpha1.Immich) []corev1.Volume {
	mlSpec := ptr.Deref(immich.Spec.MachineLearning, mediav1alpha1.MachineLearningSpec{})
	persistence := ptr.Deref(mlSpec.Persistence, mediav1alpha1.MachineLearningPersistenceSpec{})

	// Check if persistence is disabled
	if persistence.Enabled != nil && !*persistence.Enabled {
		return []corev1.Volume{
			{
				Name: "cache",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
	}

	// Persistence is enabled (default)
	pvcName := fmt.Sprintf("%s-ml-cache", immich.Name)
	if persistence.ExistingClaim != "" {
		pvcName = persistence.ExistingClaim
	}

	return []corev1.Volume{
		{
			Name: "cache",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		},
	}
}

// reconcileMLService creates or updates the ML Service using server-side apply
func (r *ImmichReconciler) reconcileMLService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-machine-learning", immich.Name)
	labels := r.getLabels(immich, "machine-learning")
	selectorLabels := r.getSelectorLabels(immich, "machine-learning")

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         immich.APIVersion,
					Kind:               immich.Kind,
					Name:               immich.Name,
					UID:                immich.UID,
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selectorLabels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       3003,
					TargetPort: intstr.FromString("http"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	return r.apply(ctx, service)
}

func (r *ImmichReconciler) reconcileMLPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	mlSpec := ptr.Deref(immich.Spec.MachineLearning, mediav1alpha1.MachineLearningSpec{})
	persistence := ptr.Deref(mlSpec.Persistence, mediav1alpha1.MachineLearningPersistenceSpec{})

	if persistence.ExistingClaim != "" {
		return nil // Using existing PVC
	}

	name := fmt.Sprintf("%s-ml-cache", immich.Name)
	labels := r.getLabels(immich, "machine-learning")

	// Check if PVC already exists - PVCs are mostly immutable
	existing := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, existing)
	if err == nil {
		// PVC exists, don't update
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	size := persistence.Size
	if size.IsZero() {
		size = resource.MustParse("10Gi")
	}

	accessModes := persistence.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	// Create new PVC with owner reference (ML cache is not as critical as library/postgres)
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: immich.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         immich.APIVersion,
					Kind:               immich.Kind,
					Name:               immich.Name,
					UID:                immich.UID,
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				},
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      accessModes,
			StorageClassName: persistence.StorageClass,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: size,
				},
			},
		},
	}

	return r.Create(ctx, pvc)
}
