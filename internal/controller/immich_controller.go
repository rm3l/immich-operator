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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

const (
	// Finalizer for Immich resources
	immichFinalizer = "media.rm3l.org/finalizer"

	// Condition types
	ConditionTypeReady       = "Ready"
	ConditionTypeProgressing = "Progressing"
	ConditionTypeDegraded    = "Degraded"
)

// ImmichReconciler reconciles a Immich object
type ImmichReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=media.rm3l.org,resources=immiches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=media.rm3l.org,resources=immiches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=media.rm3l.org,resources=immiches/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ImmichReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the Immich instance
	immich := &mediav1alpha1.Immich{}
	if err := r.Get(ctx, req.NamespacedName, immich); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Immich resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Immich")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !immich.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(immich, immichFinalizer) {
			// Run finalization logic
			if err := r.finalizeImmich(ctx, immich); err != nil {
				return ctrl.Result{}, err
			}
			// Remove finalizer
			controllerutil.RemoveFinalizer(immich, immichFinalizer)
			if err := r.Update(ctx, immich); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(immich, immichFinalizer) {
		controllerutil.AddFinalizer(immich, immichFinalizer)
		if err := r.Update(ctx, immich); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize status conditions
	if immich.Status.Conditions == nil {
		immich.Status.Conditions = []metav1.Condition{}
	}

	// Set Progressing condition
	meta.SetStatusCondition(&immich.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeProgressing,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciling",
		Message: "Reconciling Immich resources",
	})

	// Validate required images are set
	if err := r.validateImages(immich); err != nil {
		log.Error(err, "Image validation failed")
		meta.SetStatusCondition(&immich.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeDegraded,
			Status:  metav1.ConditionTrue,
			Reason:  "ImageNotConfigured",
			Message: err.Error(),
		})
		immich.Status.Ready = false
		if statusErr := r.Status().Update(ctx, immich); statusErr != nil {
			log.Error(statusErr, "Failed to update status")
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Reconcile all components
	var reconcileErr error

	// 1. Reconcile Library PVC if needed
	if immich.ShouldCreateLibraryPVC() {
		if err := r.reconcileLibraryPVC(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Library PVC")
			reconcileErr = err
		}
	}

	// 2. Reconcile Immich configuration (ConfigMap/Secret)
	if err := r.reconcileImmichConfig(ctx, immich); err != nil {
		log.Error(err, "Failed to reconcile Immich config")
		reconcileErr = err
	}

	// 3. Reconcile PostgreSQL if enabled
	if immich.IsPostgresEnabled() {
		if err := r.reconcilePostgres(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile PostgreSQL")
			reconcileErr = err
		}
	}

	// 4. Reconcile Valkey if enabled
	if immich.IsValkeyEnabled() {
		if err := r.reconcileValkey(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Valkey")
			reconcileErr = err
		}
	}

	// 5. Reconcile Machine Learning if enabled
	if immich.IsMachineLearningEnabled() {
		if err := r.reconcileMachineLearning(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Machine Learning")
			reconcileErr = err
		}
	}

	// 6. Reconcile Server if enabled
	if immich.IsServerEnabled() {
		if err := r.reconcileServer(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Server")
			reconcileErr = err
		}
	}

	// Update status
	if err := r.updateStatus(ctx, immich); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	if reconcileErr != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, reconcileErr
	}

	// Set Ready condition based on component status
	if immich.Status.Ready {
		meta.SetStatusCondition(&immich.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  "AllComponentsReady",
			Message: "All Immich components are ready",
		})
		meta.RemoveStatusCondition(&immich.Status.Conditions, ConditionTypeProgressing)
	} else {
		meta.SetStatusCondition(&immich.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  "ComponentsNotReady",
			Message: "Some Immich components are not ready",
		})
	}

	immich.Status.ObservedGeneration = immich.Generation

	if err := r.Status().Update(ctx, immich); err != nil {
		log.Error(err, "Failed to update Immich status")
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully reconciled Immich")
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// finalizeImmich handles cleanup when the Immich resource is deleted
func (r *ImmichReconciler) finalizeImmich(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing Immich")
	// Kubernetes garbage collection will handle owned resources
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImmichReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediav1alpha1.Immich{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&networkingv1.Ingress{}).
		Named("immich").
		Complete(r)
}
