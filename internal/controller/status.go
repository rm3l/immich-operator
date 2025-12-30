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
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	// Update URL from Route or Ingress
	if err := r.updateURLStatus(ctx, immich); err != nil {
		// Non-fatal error, just log it
		return err
	}

	return nil
}

// updateURLStatus updates the URL in the Immich status from Route or Ingress
func (r *ImmichReconciler) updateURLStatus(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-server", immich.Name)
	routeAPIAvailable := r.IsRouteAPIAvailable()

	// Try to get URL from Route first (if Route API is available)
	if immich.ShouldCreateRoute(routeAPIAvailable) {
		route := &unstructured.Unstructured{}
		route.SetGroupVersionKind(RouteGVK)
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, route); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		} else {
			// Extract host from Route status
			if host := getRouteHost(route); host != "" {
				// Determine protocol (check if TLS is configured)
				protocol := "http"
				if tls, found, _ := unstructured.NestedMap(route.Object, "spec", "tls"); found && tls != nil {
					protocol = "https"
				}
				immich.Status.URL = fmt.Sprintf("%s://%s", protocol, host)
				return nil
			}
		}
	}

	// Fall back to Ingress if enabled
	if immich.IsIngressEnabled() {
		ingress := &networkingv1.Ingress{}
		if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: immich.Namespace}, ingress); err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		} else {
			// Extract host from Ingress
			if host := getIngressHost(ingress); host != "" {
				// Determine protocol
				protocol := "http"
				if len(ingress.Spec.TLS) > 0 {
					protocol = "https"
				}
				immich.Status.URL = fmt.Sprintf("%s://%s", protocol, host)
				return nil
			}
		}
	}

	// No URL available yet
	immich.Status.URL = ""
	return nil
}

// getRouteHost extracts the host from an OpenShift Route
func getRouteHost(route *unstructured.Unstructured) string {
	// First try status.ingress[0].host (assigned by OpenShift)
	ingresses, found, _ := unstructured.NestedSlice(route.Object, "status", "ingress")
	if found && len(ingresses) > 0 {
		if ingressMap, ok := ingresses[0].(map[string]interface{}); ok {
			if host, ok := ingressMap["host"].(string); ok && host != "" {
				return host
			}
		}
	}

	// Fall back to spec.host (user-specified)
	host, _, _ := unstructured.NestedString(route.Object, "spec", "host")
	return host
}

// getIngressHost extracts the first host from an Ingress
func getIngressHost(ingress *networkingv1.Ingress) string {
	// First try status (load balancer assigned)
	if len(ingress.Status.LoadBalancer.Ingress) > 0 {
		lb := ingress.Status.LoadBalancer.Ingress[0]
		if lb.Hostname != "" {
			return lb.Hostname
		}
		if lb.IP != "" {
			return lb.IP
		}
	}

	// Fall back to spec.rules[0].host
	if len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != "" {
		return ingress.Spec.Rules[0].Host
	}

	return ""
}
