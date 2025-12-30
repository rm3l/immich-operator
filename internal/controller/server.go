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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// reconcileServer creates or updates the Immich Server deployment, service, and ingress/route
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

	// Check if Route API is available (OpenShift)
	routeAPIAvailable := r.IsRouteAPIAvailable()

	// Create OpenShift Route if should create (auto-detected or explicitly enabled)
	if immich.ShouldCreateRoute(routeAPIAvailable) {
		log.V(1).Info("Creating OpenShift Route (Route API available)")
		if err := r.reconcileServerRoute(ctx, immich); err != nil {
			return err
		}
	}

	// Create Server Ingress if explicitly enabled
	// Note: Ingress is only created if explicitly enabled, Route takes precedence by default on OpenShift
	if immich.IsIngressEnabled() {
		if err := r.reconcileServerIngress(ctx, immich); err != nil {
			return err
		}
	}

	return nil
}

// reconcileServerDeployment creates or updates the Server Deployment using server-side apply
func (r *ImmichReconciler) reconcileServerDeployment(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")
	selectorLabels := r.getSelectorLabels(immich, "server")

	serverSpec := ptr.Deref(immich.Spec.Server, mediav1alpha1.ServerSpec{})

	replicas := ptr.Deref(serverSpec.Replicas, 1)

	// Build environment variables
	env := r.getServerEnv(immich)
	env = append(env, serverSpec.Env...)

	// Build volume mounts and volumes
	volumeMounts := r.getServerVolumeMounts(immich)
	volumes := r.getServerVolumes(immich)

	// Add config checksum annotation if configuration exists
	annotations := make(map[string]string)
	for k, v := range serverSpec.PodAnnotations {
		annotations[k] = v
	}

	// Build container ports
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 2283,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if immich.IsMetricsEnabled() {
		ports = append(ports,
			corev1.ContainerPort{Name: "metrics-api", ContainerPort: 8081, Protocol: corev1.ProtocolTCP},
			corev1.ContainerPort{Name: "metrics-ms", ContainerPort: 8082, Protocol: corev1.ProtocolTCP},
		)
	}

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
					Labels:      r.mergeMaps(labels, serverSpec.PodLabels),
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					SecurityContext:  serverSpec.PodSecurityContext,
					ImagePullSecrets: immich.Spec.ImagePullSecrets,
					NodeSelector:     serverSpec.NodeSelector,
					Tolerations:      serverSpec.Tolerations,
					Affinity:         serverSpec.Affinity,
					InitContainers:   r.getServerInitContainers(immich),
					Containers: []corev1.Container{
						{
							Name:            "server",
							Image:           immich.GetServerImage(),
							ImagePullPolicy: serverSpec.ImagePullPolicy,
							Env:             env,
							EnvFrom:         serverSpec.EnvFrom,
							Ports:           ports,
							Resources:       serverSpec.Resources,
							SecurityContext: serverSpec.SecurityContext,
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
		},
	}

	return r.apply(ctx, deployment)
}

func (r *ImmichReconciler) getServerEnv(immich *mediav1alpha1.Immich) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	valkeySpec := ptr.Deref(immich.Spec.Valkey, mediav1alpha1.ValkeySpec{})
	postgresSpec := ptr.Deref(immich.Spec.Postgres, mediav1alpha1.PostgresSpec{})

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
		if !immich.IsValkeyEnabled() && valkeySpec.PasswordSecretRef != nil {
			env = append(env, corev1.EnvVar{
				Name: "REDIS_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: valkeySpec.PasswordSecretRef.Name,
						},
						Key: valkeySpec.PasswordSecretRef.Key,
					},
				},
			})
		}
		// Add DB index if configured (external Valkey)
		if !immich.IsValkeyEnabled() && valkeySpec.DbIndex != nil && *valkeySpec.DbIndex != 0 {
			env = append(env, corev1.EnvVar{
				Name:  "REDIS_DBINDEX",
				Value: fmt.Sprintf("%d", *valkeySpec.DbIndex),
			})
		}
	}

	// Note: Machine Learning URL is now configured via the Immich config file,
	// which is auto-generated by the operator based on CR settings.

	// Metrics
	if immich.IsMetricsEnabled() {
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
	if postgresSpec.URLSecretRef != nil {
		env = append(env, corev1.EnvVar{
			Name: "DB_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: postgresSpec.URLSecretRef.Name,
					},
					Key: postgresSpec.URLSecretRef.Key,
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

	postgresSpec := ptr.Deref(immich.Spec.Postgres, mediav1alpha1.PostgresSpec{})
	valkeySpec := ptr.Deref(immich.Spec.Valkey, mediav1alpha1.ValkeySpec{})

	// Wait for PostgreSQL
	postgresHost := fmt.Sprintf("%s-postgres", immich.Name)
	postgresPort := int32(5432)
	if !immich.IsPostgresEnabled() && postgresSpec.Host != nil && *postgresSpec.Host != "" {
		postgresHost = *postgresSpec.Host
		if postgresSpec.Port != nil && *postgresSpec.Port != 0 {
			postgresPort = *postgresSpec.Port
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
	if immich.IsValkeyEnabled() || (valkeySpec.Host != nil && *valkeySpec.Host != "") {
		valkeyHost := fmt.Sprintf("%s-valkey", immich.Name)
		valkeyPort := int32(6379)
		if !immich.IsValkeyEnabled() && valkeySpec.Host != nil && *valkeySpec.Host != "" {
			valkeyHost = *valkeySpec.Host
			if valkeySpec.Port != nil && *valkeySpec.Port != 0 {
				valkeyPort = *valkeySpec.Port
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
	immichConfig := ptr.Deref(immich.Spec.Immich, mediav1alpha1.ImmichConfig{})
	persistence := ptr.Deref(immichConfig.Persistence, mediav1alpha1.PersistenceSpec{})
	library := ptr.Deref(persistence.Library, mediav1alpha1.LibraryPersistenceSpec{})

	if (library.ExistingClaim != nil && *library.ExistingClaim != "") || immich.ShouldCreateLibraryPVC() {
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
	immichConfig := ptr.Deref(immich.Spec.Immich, mediav1alpha1.ImmichConfig{})
	persistence := ptr.Deref(immichConfig.Persistence, mediav1alpha1.PersistenceSpec{})
	library := ptr.Deref(persistence.Library, mediav1alpha1.LibraryPersistenceSpec{})

	if (library.ExistingClaim != nil && *library.ExistingClaim != "") || immich.ShouldCreateLibraryPVC() {
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
	if immich.GetConfigurationKind() == "Secret" {
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

// reconcileServerService creates or updates the Server Service using server-side apply
func (r *ImmichReconciler) reconcileServerService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")
	selectorLabels := r.getSelectorLabels(immich, "server")

	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Port:       2283,
			TargetPort: intstr.FromString("http"),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	if immich.IsMetricsEnabled() {
		ports = append(ports,
			corev1.ServicePort{Name: "metrics-api", Port: 8081, TargetPort: intstr.FromString("metrics-api"), Protocol: corev1.ProtocolTCP},
			corev1.ServicePort{Name: "metrics-ms", Port: 8082, TargetPort: intstr.FromString("metrics-ms"), Protocol: corev1.ProtocolTCP},
		)
	}

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
			Ports:    ports,
		},
	}

	return r.apply(ctx, service)
}

// reconcileServerIngress creates or updates the Server Ingress using server-side apply
func (r *ImmichReconciler) reconcileServerIngress(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")

	serverSpec := ptr.Deref(immich.Spec.Server, mediav1alpha1.ServerSpec{})
	ingress := ptr.Deref(serverSpec.Ingress, mediav1alpha1.IngressSpec{})

	// Build rules
	rules := make([]networkingv1.IngressRule, 0, len(ingress.Hosts))
	for _, host := range ingress.Hosts {
		var paths []networkingv1.HTTPIngressPath
		for _, p := range host.Paths {
			var pathType networkingv1.PathType
			pathTypeStr := ptr.Deref(p.PathType, "Prefix")
			switch pathTypeStr {
			case "Exact":
				pathType = networkingv1.PathTypeExact
			case "ImplementationSpecific":
				pathType = networkingv1.PathTypeImplementationSpecific
			default:
				pathType = networkingv1.PathTypePrefix
			}
			path := ptr.Deref(p.Path, "/")
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
		hostName := ptr.Deref(host.Host, "")
		rules = append(rules, networkingv1.IngressRule{
			Host: hostName,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}

	// Build TLS
	tls := make([]networkingv1.IngressTLS, 0, len(ingress.TLS))
	for _, t := range ingress.TLS {
		secretName := ptr.Deref(t.SecretName, "")
		tls = append(tls, networkingv1.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: secretName,
		})
	}

	ingressObj := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkingv1.SchemeGroupVersion.String(),
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   immich.Namespace,
			Labels:      labels,
			Annotations: ingress.Annotations,
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
		Spec: networkingv1.IngressSpec{
			IngressClassName: ingress.IngressClassName,
			Rules:            rules,
			TLS:              tls,
		},
	}

	return r.apply(ctx, ingressObj)
}

// reconcileServerRoute creates or updates the Server OpenShift Route using server-side apply
func (r *ImmichReconciler) reconcileServerRoute(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciling Server Route")

	name := fmt.Sprintf("%s-server", immich.Name)
	labels := r.getLabels(immich, "server")

	serverSpec := ptr.Deref(immich.Spec.Server, mediav1alpha1.ServerSpec{})
	routeSpec := ptr.Deref(serverSpec.Route, mediav1alpha1.RouteSpec{})

	// Merge labels
	routeLabels := r.mergeMaps(labels, routeSpec.Labels)

	// Build the Route object as unstructured since we don't want to import OpenShift types
	// This keeps the operator compatible with both vanilla Kubernetes and OpenShift
	route := map[string]interface{}{
		"apiVersion": "route.openshift.io/v1",
		"kind":       "Route",
		"metadata": map[string]interface{}{
			"name":        name,
			"namespace":   immich.Namespace,
			"labels":      routeLabels,
			"annotations": routeSpec.Annotations,
			"ownerReferences": []map[string]interface{}{
				{
					"apiVersion":         immich.APIVersion,
					"kind":               immich.Kind,
					"name":               immich.Name,
					"uid":                string(immich.UID),
					"controller":         true,
					"blockOwnerDeletion": true,
				},
			},
		},
		"spec": map[string]interface{}{
			"to": map[string]interface{}{
				"kind":   "Service",
				"name":   name,
				"weight": int64(100),
			},
			"port": map[string]interface{}{
				"targetPort": "http",
			},
			"wildcardPolicy": ptr.Deref(routeSpec.WildcardPolicy, "None"),
		},
	}

	// Add host if specified
	if routeSpec.Host != nil && *routeSpec.Host != "" {
		route["spec"].(map[string]interface{})["host"] = *routeSpec.Host
	}

	// Add path if specified
	if routeSpec.Path != nil && *routeSpec.Path != "" && *routeSpec.Path != "/" {
		route["spec"].(map[string]interface{})["path"] = *routeSpec.Path
	}

	// Add TLS configuration if specified
	if routeSpec.TLS != nil {
		tlsConfig := map[string]interface{}{
			"termination":                   ptr.Deref(routeSpec.TLS.Termination, "edge"),
			"insecureEdgeTerminationPolicy": ptr.Deref(routeSpec.TLS.InsecureEdgeTerminationPolicy, "Redirect"),
		}

		if routeSpec.TLS.Certificate != nil && *routeSpec.TLS.Certificate != "" {
			tlsConfig["certificate"] = *routeSpec.TLS.Certificate
		}
		if routeSpec.TLS.Key != nil && *routeSpec.TLS.Key != "" {
			tlsConfig["key"] = *routeSpec.TLS.Key
		}
		if routeSpec.TLS.CACertificate != nil && *routeSpec.TLS.CACertificate != "" {
			tlsConfig["caCertificate"] = *routeSpec.TLS.CACertificate
		}
		if routeSpec.TLS.DestinationCACertificate != nil && *routeSpec.TLS.DestinationCACertificate != "" {
			tlsConfig["destinationCACertificate"] = *routeSpec.TLS.DestinationCACertificate
		}

		route["spec"].(map[string]interface{})["tls"] = tlsConfig
	}

	// Convert to unstructured for SSA
	unstructuredRoute := &unstructured.Unstructured{Object: route}

	return r.apply(ctx, unstructuredRoute)
}
