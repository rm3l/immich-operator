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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// reconcileServer creates or updates the Immich Server deployment, service, and ingress
func (r *ImmichReconciler) reconcileServer(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Server")

	// Create Server Deployment
	if err := r.reconcileServerDeployment(ctx, immich); err != nil {
		return err
	}

	// Create Server Service
	if err := r.reconcileServerService(ctx, immich); err != nil {
		return err
	}

	// Create Server Ingress if enabled
	if immich.Spec.Server.Ingress.Enabled {
		if err := r.reconcileServerIngress(ctx, immich); err != nil {
			return err
		}
	}

	return nil
}

func (r *ImmichReconciler) reconcileServerDeployment(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")
	selectorLabels := r.getSelectorLabels(immich, "server")

	replicas := int32(1)
	if immich.Spec.Server.Replicas != nil {
		replicas = *immich.Spec.Server.Replicas
	}

	// Build environment variables
	env := r.getServerEnv(immich)
	env = append(env, immich.Spec.Server.Env...)

	// Build volume mounts and volumes
	volumeMounts := r.getServerVolumeMounts(immich)
	volumes := r.getServerVolumes(immich)

	// Add config checksum annotation if configuration exists
	annotations := make(map[string]string)
	for k, v := range immich.Spec.Server.PodAnnotations {
		annotations[k] = v
	}

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

	// Build container ports
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 2283,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if immich.Spec.Immich.Metrics.Enabled {
		ports = append(ports,
			corev1.ContainerPort{Name: "metrics-api", ContainerPort: 8081, Protocol: corev1.ProtocolTCP},
			corev1.ContainerPort{Name: "metrics-ms", ContainerPort: 8082, Protocol: corev1.ProtocolTCP},
		)
	}

	return r.createOrUpdate(ctx, deployment, func() error {
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.mergeMaps(labels, immich.Spec.Server.PodLabels),
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					SecurityContext:  immich.Spec.Server.PodSecurityContext,
					ImagePullSecrets: immich.Spec.ImagePullSecrets,
					NodeSelector:     immich.Spec.Server.NodeSelector,
					Tolerations:      immich.Spec.Server.Tolerations,
					Affinity:         immich.Spec.Server.Affinity,
					InitContainers:   r.getServerInitContainers(immich),
					Containers: []corev1.Container{
						{
							Name:            "server",
							Image:           immich.GetServerImage(),
							ImagePullPolicy: immich.Spec.Server.ImagePullPolicy,
							Env:             env,
							EnvFrom:         immich.Spec.Server.EnvFrom,
							Ports:           ports,
							Resources:       immich.Spec.Server.Resources,
							SecurityContext: immich.Spec.Server.SecurityContext,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/server/ping",
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
										Path: "/api/server/ping",
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
										Path: "/api/server/ping",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								FailureThreshold:    30,
							},
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		}
		return nil
	})
}

func (r *ImmichReconciler) getServerEnv(immich *mediav1alpha1.Immich) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	// Redis/Valkey connection - uses helper to determine built-in vs external
	valkeyHost := immich.GetValkeyHost()
	if valkeyHost != "" {
		env = append(env, corev1.EnvVar{
			Name:  "REDIS_HOSTNAME",
			Value: valkeyHost,
		})
		env = append(env, corev1.EnvVar{
			Name:  "REDIS_PORT",
			Value: fmt.Sprintf("%d", immich.GetValkeyPort()),
		})
		// Add password if configured (external Valkey)
		if !immich.IsValkeyEnabled() && immich.Spec.Valkey.PasswordSecretRef != nil {
			env = append(env, corev1.EnvVar{
				Name: "REDIS_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: immich.Spec.Valkey.PasswordSecretRef.Name,
						},
						Key: immich.Spec.Valkey.PasswordSecretRef.Key,
					},
				},
			})
		}
		// Add DB index if configured (external Valkey)
		if !immich.IsValkeyEnabled() && immich.Spec.Valkey.DbIndex != 0 {
			env = append(env, corev1.EnvVar{
				Name:  "REDIS_DBINDEX",
				Value: fmt.Sprintf("%d", immich.Spec.Valkey.DbIndex),
			})
		}
	}

	// Note: Machine Learning URL is now configured via the Immich config file,
	// which is auto-generated by the operator based on CR settings.

	// Metrics
	if immich.Spec.Immich.Metrics.Enabled {
		env = append(env, corev1.EnvVar{
			Name:  "IMMICH_TELEMETRY_INCLUDE",
			Value: "all",
		})
	}

	// Config file path - always set since we always generate a config
	env = append(env, corev1.EnvVar{
		Name:  "IMMICH_CONFIG_FILE",
		Value: "/config/immich-config.yaml",
	})

	// Database configuration - uses helper methods to determine built-in vs external
	if immich.Spec.Postgres.URLSecretRef != nil {
		env = append(env, corev1.EnvVar{
			Name: "DB_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: immich.Spec.Postgres.URLSecretRef.Name,
					},
					Key: immich.Spec.Postgres.URLSecretRef.Key,
				},
			},
		})
	} else {
		// Use helper methods which handle built-in vs external automatically
		env = append(env, corev1.EnvVar{
			Name:  "DB_HOSTNAME",
			Value: immich.GetPostgresHost(),
		})
		env = append(env, corev1.EnvVar{
			Name:  "DB_PORT",
			Value: fmt.Sprintf("%d", immich.GetPostgresPort()),
		})
		env = append(env, corev1.EnvVar{
			Name:  "DB_DATABASE_NAME",
			Value: immich.GetPostgresDatabase(),
		})
		env = append(env, corev1.EnvVar{
			Name:  "DB_USERNAME",
			Value: immich.GetPostgresUsername(),
		})

		// Use secret reference (user-provided or auto-generated for built-in PostgreSQL)
		secretRef := r.getPostgresPasswordSecretRef(immich)
		env = append(env, corev1.EnvVar{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretRef.Name,
					},
					Key: secretRef.Key,
				},
			},
		})
	}

	return env
}

// getServerInitContainers returns init containers that wait for dependencies
func (r *ImmichReconciler) getServerInitContainers(immich *mediav1alpha1.Immich) []corev1.Container {
	initContainers := []corev1.Container{}

	// Get init container image from environment variable
	initImage := mediav1alpha1.GetImmichInitContainerImage()
	if initImage == "" {
		return initContainers // Skip init containers if no image is configured
	}

	// Wait for PostgreSQL
	postgresHost := fmt.Sprintf("%s-postgres", immich.Name)
	postgresPort := int32(5432)
	if !immich.IsPostgresEnabled() && immich.Spec.Postgres.Host != "" {
		postgresHost = immich.Spec.Postgres.Host
		if immich.Spec.Postgres.Port != 0 {
			postgresPort = immich.Spec.Postgres.Port
		}
	}

	initContainers = append(initContainers, corev1.Container{
		Name:  "wait-for-postgres",
		Image: initImage,
		Command: []string{
			"sh", "-c",
			fmt.Sprintf(`echo "Waiting for PostgreSQL at %s:%d..."
until nc -z -w2 %s %d; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done
echo "PostgreSQL is up"`, postgresHost, postgresPort, postgresHost, postgresPort),
		},
	})

	// Wait for Valkey/Redis
	if immich.IsValkeyEnabled() || immich.Spec.Valkey.Host != "" {
		valkeyHost := fmt.Sprintf("%s-valkey", immich.Name)
		valkeyPort := int32(6379)
		if !immich.IsValkeyEnabled() && immich.Spec.Valkey.Host != "" {
			valkeyHost = immich.Spec.Valkey.Host
			if immich.Spec.Valkey.Port != 0 {
				valkeyPort = immich.Spec.Valkey.Port
			}
		}

		initContainers = append(initContainers, corev1.Container{
			Name:  "wait-for-valkey",
			Image: initImage,
			Command: []string{
				"sh", "-c",
				fmt.Sprintf(`echo "Waiting for Valkey at %s:%d..."
until nc -z -w2 %s %d; do
  echo "Valkey is unavailable - sleeping"
  sleep 2
done
echo "Valkey is up"`, valkeyHost, valkeyPort, valkeyHost, valkeyPort),
			},
		})
	}

	return initContainers
}

func (r *ImmichReconciler) getServerVolumeMounts(immich *mediav1alpha1.Immich) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}

	// Library mount (when using existing PVC or operator-managed PVC)
	if immich.Spec.Immich.Persistence.Library.ExistingClaim != "" || immich.ShouldCreateLibraryPVC() {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "library",
			MountPath: "/data",
		})
	}

	// Config mount - always mounted since we always generate a config
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "config",
		MountPath: "/config",
		ReadOnly:  true,
	})

	return mounts
}

func (r *ImmichReconciler) getServerVolumes(immich *mediav1alpha1.Immich) []corev1.Volume {
	volumes := []corev1.Volume{}

	// Library volume (when using existing PVC or operator-managed PVC)
	if immich.Spec.Immich.Persistence.Library.ExistingClaim != "" || immich.ShouldCreateLibraryPVC() {
		volumes = append(volumes, corev1.Volume{
			Name: "library",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: immich.GetLibraryPVCName(),
				},
			},
		})
	}

	// Config volume - always created since we always generate a config
	configName := fmt.Sprintf("%s-immich-config", immich.Name)
	if immich.Spec.Immich.ConfigurationKind == "Secret" {
		volumes = append(volumes, corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: configName,
				},
			},
		})
	} else {
		volumes = append(volumes, corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configName,
					},
				},
			},
		})
	}

	return volumes
}

func (r *ImmichReconciler) reconcileServerService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")
	selectorLabels := r.getSelectorLabels(immich, "server")

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

	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Port:       2283,
			TargetPort: intstr.FromString("http"),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if immich.Spec.Immich.Metrics.Enabled {
		ports = append(ports,
			corev1.ServicePort{Name: "metrics-api", Port: 8081, TargetPort: intstr.FromString("metrics-api"), Protocol: corev1.ProtocolTCP},
			corev1.ServicePort{Name: "metrics-ms", Port: 8082, TargetPort: intstr.FromString("metrics-ms"), Protocol: corev1.ProtocolTCP},
		)
	}

	return r.createOrUpdate(ctx, service, func() error {
		service.Spec = corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selectorLabels,
			Ports:    ports,
		}
		return nil
	})
}

func (r *ImmichReconciler) reconcileServerIngress(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   immich.Namespace,
			Labels:      labels,
			Annotations: immich.Spec.Server.Ingress.Annotations,
		},
	}

	if err := controllerutil.SetControllerReference(immich, ingress, r.Scheme); err != nil {
		return err
	}

	return r.createOrUpdate(ctx, ingress, func() error {
		// Build rules
		var rules []networkingv1.IngressRule
		for _, host := range immich.Spec.Server.Ingress.Hosts {
			var paths []networkingv1.HTTPIngressPath
			for _, p := range host.Paths {
				var pathType networkingv1.PathType
				switch p.PathType {
				case "Exact":
					pathType = networkingv1.PathTypeExact
				case "ImplementationSpecific":
					pathType = networkingv1.PathTypeImplementationSpecific
				default:
					pathType = networkingv1.PathTypePrefix
				}
				path := p.Path
				if path == "" {
					path = "/"
				}
				paths = append(paths, networkingv1.HTTPIngressPath{
					Path:     path,
					PathType: &pathType,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: name,
							Port: networkingv1.ServiceBackendPort{
								Name: "http",
							},
						},
					},
				})
			}
			rules = append(rules, networkingv1.IngressRule{
				Host: host.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: paths,
					},
				},
			})
		}

		// Build TLS
		var tls []networkingv1.IngressTLS
		for _, t := range immich.Spec.Server.Ingress.TLS {
			tls = append(tls, networkingv1.IngressTLS{
				Hosts:      t.Hosts,
				SecretName: t.SecretName,
			})
		}

		ingress.Spec = networkingv1.IngressSpec{
			IngressClassName: immich.Spec.Server.Ingress.IngressClassName,
			Rules:            rules,
			TLS:              tls,
		}
		return nil
	})
}
