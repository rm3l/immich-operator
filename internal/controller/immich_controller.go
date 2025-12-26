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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

const (
	// Finalizer for Immich resources
	immichFinalizer = "media.rm3l.org/finalizer"

	// Condition types
	ConditionTypeReady       = "Ready"
	ConditionTypeProgressing = "Progressing"
	ConditionTypeDegraded    = "Degraded"

	// Labels
	labelApp       = "app.kubernetes.io/name"
	labelInstance  = "app.kubernetes.io/instance"
	labelComponent = "app.kubernetes.io/component"
	labelManagedBy = "app.kubernetes.io/managed-by"
	labelPartOf    = "app.kubernetes.io/part-of"
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

	// Reconcile all components
	var reconcileErr error

	// 1. Reconcile Immich configuration (ConfigMap/Secret)
	if err := r.reconcileImmichConfig(ctx, immich); err != nil {
		log.Error(err, "Failed to reconcile Immich config")
		reconcileErr = err
	}

	// 2. Reconcile Valkey if enabled
	if immich.IsValkeyEnabled() {
		if err := r.reconcileValkey(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Valkey")
			reconcileErr = err
		}
	}

	// 3. Reconcile Machine Learning if enabled
	if immich.IsMachineLearningEnabled() {
		if err := r.reconcileMachineLearning(ctx, immich); err != nil {
			log.Error(err, "Failed to reconcile Machine Learning")
			reconcileErr = err
		}
	}

	// 4. Reconcile Server if enabled
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
	immich.Status.Version = immich.GetImageTag()

	if err := r.Status().Update(ctx, immich); err != nil {
		log.Error(err, "Failed to update Immich status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled Immich")
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// finalizeImmich handles cleanup when the Immich resource is deleted
func (r *ImmichReconciler) finalizeImmich(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing Immich")
	// Kubernetes garbage collection will handle owned resources
	return nil
}

// reconcileImmichConfig creates or updates the Immich configuration ConfigMap or Secret
func (r *ImmichReconciler) reconcileImmichConfig(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)

	if immich.Spec.Immich.Configuration == nil {
		log.V(1).Info("No Immich configuration specified, skipping ConfigMap/Secret creation")
		return nil
	}

	configName := fmt.Sprintf("%s-immich-config", immich.Name)

	// Convert configuration to YAML
	configData, err := yaml.Marshal(immich.Spec.Immich.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal immich configuration: %w", err)
	}

	labels := r.getLabels(immich, "config")

	if immich.Spec.Immich.ConfigurationKind == "Secret" {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configName,
				Namespace: immich.Namespace,
				Labels:    labels,
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"immich-config.yaml": string(configData),
			},
		}

		if err := controllerutil.SetControllerReference(immich, secret, r.Scheme); err != nil {
			return err
		}

		return r.createOrUpdate(ctx, secret, func() error {
			secret.StringData = map[string]string{
				"immich-config.yaml": string(configData),
			}
			return nil
		})
	}

	// Default to ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: immich.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"immich-config.yaml": string(configData),
		},
	}

	if err := controllerutil.SetControllerReference(immich, configMap, r.Scheme); err != nil {
		return err
	}

	return r.createOrUpdate(ctx, configMap, func() error {
		configMap.Data = map[string]string{
			"immich-config.yaml": string(configData),
		}
		return nil
	})
}

// reconcileValkey creates or updates the Valkey (Redis) deployment and service
func (r *ImmichReconciler) reconcileValkey(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Valkey")

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
					ImagePullSecrets: immich.Spec.Image.PullSecrets,
					NodeSelector:     immich.Spec.Valkey.NodeSelector,
					Tolerations:      immich.Spec.Valkey.Tolerations,
					Affinity:         immich.Spec.Valkey.Affinity,
					Containers: []corev1.Container{
						{
							Name:            "valkey",
							Image:           immich.GetValkeyImage(),
							ImagePullPolicy: immich.GetImagePullPolicy(),
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

// reconcileMachineLearning creates or updates the Machine Learning deployment and service
func (r *ImmichReconciler) reconcileMachineLearning(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Machine Learning")

	// Create ML Deployment
	if err := r.reconcileMLDeployment(ctx, immich); err != nil {
		return err
	}

	// Create ML Service
	if err := r.reconcileMLService(ctx, immich); err != nil {
		return err
	}

	// Create ML PVC if persistence is enabled
	persistenceEnabled := immich.Spec.MachineLearning.Persistence.Enabled
	if persistenceEnabled == nil || *persistenceEnabled {
		if err := r.reconcileMLPVC(ctx, immich); err != nil {
			return err
		}
	}

	return nil
}

func (r *ImmichReconciler) reconcileMLDeployment(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-machine-learning", immich.Name)
	labels := r.getLabels(immich, "machine-learning")
	selectorLabels := r.getSelectorLabels(immich, "machine-learning")

	replicas := int32(1)
	if immich.Spec.MachineLearning.Replicas != nil {
		replicas = *immich.Spec.MachineLearning.Replicas
	}

	env := []corev1.EnvVar{
		{Name: "TRANSFORMERS_CACHE", Value: "/cache"},
		{Name: "HF_XET_CACHE", Value: "/cache/huggingface-xet"},
		{Name: "MPLCONFIGDIR", Value: "/cache/matplotlib-config"},
	}
	env = append(env, immich.Spec.MachineLearning.Env...)

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
			Replicas: ptr.To(replicas),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.mergeMaps(labels, immich.Spec.MachineLearning.PodLabels),
					Annotations: immich.Spec.MachineLearning.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					SecurityContext:  immich.Spec.MachineLearning.PodSecurityContext,
					ImagePullSecrets: immich.Spec.Image.PullSecrets,
					NodeSelector:     immich.Spec.MachineLearning.NodeSelector,
					Tolerations:      immich.Spec.MachineLearning.Tolerations,
					Affinity:         immich.Spec.MachineLearning.Affinity,
					Containers: []corev1.Container{
						{
							Name:            "machine-learning",
							Image:           immich.GetMachineLearningImage(),
							ImagePullPolicy: immich.GetImagePullPolicy(),
							Env:             env,
							EnvFrom:         immich.Spec.MachineLearning.EnvFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 3003,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources:       immich.Spec.MachineLearning.Resources,
							SecurityContext: immich.Spec.MachineLearning.SecurityContext,
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
		}
		return nil
	})
}

func (r *ImmichReconciler) getMLVolumeMounts(immich *mediav1alpha1.Immich) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "cache",
			MountPath: "/cache",
		},
	}
}

func (r *ImmichReconciler) getMLVolumes(immich *mediav1alpha1.Immich) []corev1.Volume {
	persistenceEnabled := immich.Spec.MachineLearning.Persistence.Enabled
	if persistenceEnabled != nil && !*persistenceEnabled {
		return []corev1.Volume{
			{
				Name: "cache",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
	}

	pvcName := fmt.Sprintf("%s-ml-cache", immich.Name)
	if immich.Spec.MachineLearning.Persistence.ExistingClaim != "" {
		pvcName = immich.Spec.MachineLearning.Persistence.ExistingClaim
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

func (r *ImmichReconciler) reconcileMLService(ctx context.Context, immich *mediav1alpha1.Immich) error {
	name := fmt.Sprintf("%s-machine-learning", immich.Name)
	labels := r.getLabels(immich, "machine-learning")
	selectorLabels := r.getSelectorLabels(immich, "machine-learning")

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
					Name:       "http",
					Port:       3003,
					TargetPort: intstr.FromString("http"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}
		return nil
	})
}

func (r *ImmichReconciler) reconcileMLPVC(ctx context.Context, immich *mediav1alpha1.Immich) error {
	if immich.Spec.MachineLearning.Persistence.ExistingClaim != "" {
		return nil // Using existing PVC
	}

	name := fmt.Sprintf("%s-ml-cache", immich.Name)
	labels := r.getLabels(immich, "machine-learning")

	size := immich.Spec.MachineLearning.Persistence.Size
	if size.IsZero() {
		size = resource.MustParse("10Gi")
	}

	accessModes := immich.Spec.MachineLearning.Persistence.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
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
		StorageClassName: immich.Spec.MachineLearning.Persistence.StorageClass,
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: size,
			},
		},
	}

	return r.Create(ctx, pvc)
}

// reconcileServer creates or updates the Immich Server deployment, service, and ingress
func (r *ImmichReconciler) reconcileServer(ctx context.Context, immich *mediav1alpha1.Immich) error {
	log := logf.FromContext(ctx)
	log.Info("Reconciling Server")

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
					ImagePullSecrets: immich.Spec.Image.PullSecrets,
					NodeSelector:     immich.Spec.Server.NodeSelector,
					Tolerations:      immich.Spec.Server.Tolerations,
					Affinity:         immich.Spec.Server.Affinity,
					Containers: []corev1.Container{
						{
							Name:            "server",
							Image:           immich.GetServerImage(),
							ImagePullPolicy: immich.GetImagePullPolicy(),
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

	// Redis/Valkey hostname
	if immich.IsValkeyEnabled() {
		env = append(env, corev1.EnvVar{
			Name:  "REDIS_HOSTNAME",
			Value: fmt.Sprintf("%s-valkey", immich.Name),
		})
	}

	// Machine Learning URL
	if immich.IsMachineLearningEnabled() {
		env = append(env, corev1.EnvVar{
			Name:  "IMMICH_MACHINE_LEARNING_URL",
			Value: fmt.Sprintf("http://%s-machine-learning:3003", immich.Name),
		})
	}

	// Metrics
	if immich.Spec.Immich.Metrics.Enabled {
		env = append(env, corev1.EnvVar{
			Name:  "IMMICH_TELEMETRY_INCLUDE",
			Value: "all",
		})
	}

	// Config file path
	if immich.Spec.Immich.Configuration != nil {
		env = append(env, corev1.EnvVar{
			Name:  "IMMICH_CONFIG_FILE",
			Value: "/config/immich-config.yaml",
		})
	}

	// Database configuration
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
		if immich.Spec.Postgres.Host != "" {
			env = append(env, corev1.EnvVar{
				Name:  "DB_HOSTNAME",
				Value: immich.Spec.Postgres.Host,
			})
		}
		port := immich.Spec.Postgres.Port
		if port == 0 {
			port = 5432
		}
		env = append(env, corev1.EnvVar{
			Name:  "DB_PORT",
			Value: fmt.Sprintf("%d", port),
		})

		database := immich.Spec.Postgres.Database
		if database == "" {
			database = "immich"
		}
		env = append(env, corev1.EnvVar{
			Name:  "DB_DATABASE_NAME",
			Value: database,
		})

		username := immich.Spec.Postgres.Username
		if username == "" {
			username = "immich"
		}
		env = append(env, corev1.EnvVar{
			Name:  "DB_USERNAME",
			Value: username,
		})

		if immich.Spec.Postgres.PasswordSecretRef != nil {
			env = append(env, corev1.EnvVar{
				Name: "DB_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: immich.Spec.Postgres.PasswordSecretRef.Name,
						},
						Key: immich.Spec.Postgres.PasswordSecretRef.Key,
					},
				},
			})
		} else if immich.Spec.Postgres.Password != "" {
			env = append(env, corev1.EnvVar{
				Name:  "DB_PASSWORD",
				Value: immich.Spec.Postgres.Password,
			})
		}
	}

	return env
}

func (r *ImmichReconciler) getServerVolumeMounts(immich *mediav1alpha1.Immich) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{}

	// Library mount
	if immich.Spec.Immich.Persistence.Library.ExistingClaim != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "library",
			MountPath: "/usr/src/app/upload",
		})
	}

	// Config mount
	if immich.Spec.Immich.Configuration != nil {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "config",
			MountPath: "/config",
			ReadOnly:  true,
		})
	}

	return mounts
}

func (r *ImmichReconciler) getServerVolumes(immich *mediav1alpha1.Immich) []corev1.Volume {
	volumes := []corev1.Volume{}

	// Library volume
	if immich.Spec.Immich.Persistence.Library.ExistingClaim != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "library",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: immich.Spec.Immich.Persistence.Library.ExistingClaim,
				},
			},
		})
	}

	// Config volume
	if immich.Spec.Immich.Configuration != nil {
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
				pathType := networkingv1.PathTypePrefix
				if p.PathType == "Exact" {
					pathType = networkingv1.PathTypeExact
				} else if p.PathType == "ImplementationSpecific" {
					pathType = networkingv1.PathTypeImplementationSpecific
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

	// Overall ready status
	immich.Status.Ready = immich.Status.ServerReady &&
		immich.Status.MachineLearningReady &&
		immich.Status.ValkeyReady

	return nil
}

// Helper functions

func (r *ImmichReconciler) getLabels(immich *mediav1alpha1.Immich, component string) map[string]string {
	return map[string]string{
		labelApp:       "immich",
		labelInstance:  immich.Name,
		labelComponent: component,
		labelManagedBy: "immich-operator",
		labelPartOf:    "immich",
	}
}

func (r *ImmichReconciler) getSelectorLabels(immich *mediav1alpha1.Immich, component string) map[string]string {
	return map[string]string{
		labelApp:       "immich",
		labelInstance:  immich.Name,
		labelComponent: component,
	}
}

func (r *ImmichReconciler) mergeMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

func (r *ImmichReconciler) createOrUpdate(ctx context.Context, obj client.Object, mutate func() error) error {
	log := logf.FromContext(ctx)

	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, key, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object doesn't exist, create it
			if err := mutate(); err != nil {
				return err
			}
			log.Info("Creating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
			return r.Create(ctx, obj)
		}
		return err
	}

	// Object exists, update it
	obj.SetResourceVersion(existing.GetResourceVersion())
	if err := mutate(); err != nil {
		return err
	}

	// Check if update is needed
	if equality.Semantic.DeepEqual(existing, obj) {
		return nil
	}

	log.Info("Updating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
	return r.Update(ctx, obj)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImmichReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediav1alpha1.Immich{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&networkingv1.Ingress{}).
		Named("immich").
		Complete(r)
}
