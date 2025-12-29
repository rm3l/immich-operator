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

// reconcilePostgres creates or updates the PostgreSQL StatefulSet and service
func (r *ImmichReconciler) reconcilePostgres(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling PostgreSQL")

	// Create PostgreSQL credentials secret (if needed)
	if err := r.reconcilePostgresCredentials(ctx, immich); err != nil {
		return err
	}

	// Create PostgreSQL StatefulSet (with VolumeClaimTemplate for data persistence)
	if err := r.reconcilePostgresStatefulSet(ctx, immich); err != nil {
		return err
	}

	// Create PostgreSQL Service
	if err := r.reconcilePostgresService(ctx, immich); err != nil {
		return err
	}

	return nil
}

// reconcilePostgresCredentials creates a secret with PostgreSQL credentials if not provided.
// Note: The credentials secret does NOT have an owner reference to persist alongside the PVC.
func (r *ImmichReconciler) reconcilePostgresCredentials(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)

	postgresSpec := ptr.Deref(immich.Spec.Postgres, mediav1alpha1.PostgresSpec{})

	// Skip if user provided explicit credentials
	if postgresSpec.PasswordSecretRef != nil {
		log.V(1).Info("Using user-provided PostgreSQL credentials")
		return nil
	}

	// Generate credentials secret for built-in PostgreSQL
	secretName := fmt.Sprintf("%s-postgres-credentials", immich.Name)
	labels := r.getLabels(immich, "postgres")

	// Check if secret already exists - reuse it if so
	existing := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: immich.Namespace}, existing)
	if err == nil {
		// Secret exists, reuse it (credentials must stay consistent with the database)
		log.V(1).Info("PostgreSQL credentials secret already exists, reusing", "name", secretName)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// Generate random password
	password, err := generateRandomPassword(32)
	if err != nil {
		return fmt.Errorf("failed to generate PostgreSQL password: %w", err)
	}

	// Create secret without owner reference for data safety
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"password": []byte(password),
			"username": []byte(immich.GetPostgresUsername()),
			"database": []byte(immich.GetPostgresDatabase()),
		},
	}

	// Note: We intentionally do NOT set owner reference here.
	// This ensures the credentials persist when the Immich CR is deleted,
	// staying consistent with the PostgreSQL PVC data.

	log.Info("Creating PostgreSQL credentials secret (no owner reference for data safety)", "name", secretName)
	return r.Create(ctx, secret)
}

// getPostgresPasswordSecretRef returns the secret reference for PostgreSQL password
// Returns generated secret name if no explicit credentials are provided
func (r *ImmichReconciler) getPostgresPasswordSecretRef(immich *mediav1alpha1.Immich) *mediav1alpha1.SecretKeySelector {
	postgresSpec := ptr.Deref(immich.Spec.Postgres, mediav1alpha1.PostgresSpec{})
	if postgresSpec.PasswordSecretRef != nil {
		return postgresSpec.PasswordSecretRef
	}
	// Use generated credentials secret
	return &mediav1alpha1.SecretKeySelector{
		Name: fmt.Sprintf("%s-postgres-credentials", immich.Name),
		Key:  "password",
	}
}

// reconcilePostgresStatefulSet creates or updates the PostgreSQL StatefulSet using server-side apply
func (r *ImmichReconciler) reconcilePostgresStatefulSet(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-postgres", immich.Name)
	labels := r.getLabels(immich, "postgres")

	postgresSpec := ptr.Deref(immich.Spec.Postgres, mediav1alpha1.PostgresSpec{})
	persistence := ptr.Deref(postgresSpec.Persistence, mediav1alpha1.PostgresPersistenceSpec{})

	image := immich.GetPostgresImage()
	if image == "" {
		return fmt.Errorf("PostgreSQL image not configured: set spec.postgres.image or RELATED_IMAGE_postgres environment variable")
	}

	// Get password from secret (user-provided or auto-generated)
	secretRef := r.getPostgresPasswordSecretRef(immich)
	passwordEnvVar := corev1.EnvVar{
		Name: "POSTGRES_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretRef.Name,
				},
				Key: secretRef.Key,
			},
		},
	}

	env := []corev1.EnvVar{
		{Name: "POSTGRES_USER", Value: immich.GetPostgresUsername()},
		{Name: "POSTGRES_DB", Value: immich.GetPostgresDatabase()},
		{Name: "POSTGRES_INITDB_ARGS", Value: "--data-checksums"},
		passwordEnvVar,
	}

	// Build volume mounts
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: "/var/lib/postgresql/data",
		},
	}

	// Build volumes - only needed if using an existing claim
	var volumes []corev1.Volume
	if persistence.ExistingClaim != nil && *persistence.ExistingClaim != "" {
		volumes = []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: *persistence.ExistingClaim,
					},
				},
			},
		}
	}

	// Build VolumeClaimTemplate for automatic PVC management (if not using existing claim)
	var volumeClaimTemplates []corev1.PersistentVolumeClaim
	if persistence.ExistingClaim == nil || *persistence.ExistingClaim == "" {
		size := resource.MustParse("10Gi")
		if persistence.Size != nil && !persistence.Size.IsZero() {
			size = *persistence.Size
		}

		accessModes := persistence.AccessModes
		if len(accessModes) == 0 {
			accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		}

		volumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "data",
					Labels: labels,
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
			},
		}
	}

	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "StatefulSet",
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
		Spec: appsv1.StatefulSetSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName:          name,
			VolumeClaimTemplates: volumeClaimTemplates,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: postgresSpec.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: immich.Spec.ImagePullSecrets,
					SecurityContext:  postgresSpec.PodSecurityContext,
					NodeSelector:     postgresSpec.NodeSelector,
					Tolerations:      postgresSpec.Tolerations,
					Affinity:         postgresSpec.Affinity,
					Volumes:          volumes,
					Containers: []corev1.Container{
						{
							Name:            "postgres",
							Image:           image,
							ImagePullPolicy: postgresSpec.ImagePullPolicy,
							Env:             env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "postgres",
									ContainerPort: 5432,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts:    volumeMounts,
							Resources:       postgresSpec.Resources,
							SecurityContext: postgresSpec.SecurityContext,
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"pg_isready", "-U", immich.GetPostgresUsername(), "-d", immich.GetPostgresDatabase()},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"pg_isready", "-U", immich.GetPostgresUsername(), "-d", immich.GetPostgresDatabase()},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}

	return r.apply(ctx, sts)
}

// reconcilePostgresService creates or updates the PostgreSQL Service using server-side apply
func (r *ImmichReconciler) reconcilePostgresService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-postgres", immich.Name)
	labels := r.getLabels(immich, "postgres")

	svc := &corev1.Service{
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
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "postgres",
					Port:       5432,
					TargetPort: intstr.FromString("postgres"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	return r.apply(ctx, svc)
}
