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

package v1alpha1

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Environment variable names for disconnected/air-gapped environments
// These follow the RELATED_IMAGE_* pattern used by OpenShift OLM
const (
	EnvRelatedImageImmich              = "RELATED_IMAGE_immich"
	EnvRelatedImageMachineLearning     = "RELATED_IMAGE_machineLearning"
	EnvRelatedImageValkey              = "RELATED_IMAGE_valkey"
	EnvRelatedImagePostgres            = "RELATED_IMAGE_postgres"
	EnvRelatedImageImmichInitContainer = "RELATED_IMAGE_immich_initContainer"
)

// ImmichSpec defines the desired state of Immich.
type ImmichSpec struct {
	// ImagePullSecrets are the secrets used to pull images from private registries
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Immich shared configuration
	// +optional
	Immich *ImmichConfig `json:"immich,omitempty"`

	// Server component configuration
	// +optional
	Server *ServerSpec `json:"server,omitempty"`

	// MachineLearning component configuration
	// +optional
	MachineLearning *MachineLearningSpec `json:"machineLearning,omitempty"`

	// Valkey (Redis) component configuration
	// +optional
	Valkey *ValkeySpec `json:"valkey,omitempty"`

	// PostgreSQL database configuration
	// +optional
	Postgres *PostgresSpec `json:"postgres,omitempty"`
}

// ImmichConfig defines shared Immich configuration.
type ImmichConfig struct {
	// Metrics configuration
	// +optional
	Metrics *MetricsSpec `json:"metrics,omitempty"`

	// Persistence configuration for photo library
	// +optional
	Persistence *PersistenceSpec `json:"persistence,omitempty"`

	// Configuration is immich-config.yaml converted to raw YAML
	// ref: https://immich.app/docs/install/config-file/
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// ConfigurationKind sets the resource Kind to store configuration in.
	// Must be either ConfigMap or Secret. Defaults to ConfigMap.
	// +kubebuilder:validation:Enum=ConfigMap;Secret
	// +optional
	ConfigurationKind *string `json:"configurationKind,omitempty"`
}

// ConfigurationSpec holds the raw Immich configuration
// +kubebuilder:pruning:PreserveUnknownFields
type ConfigurationSpec struct {
	// Trash configuration
	// +optional
	Trash *TrashConfig `json:"trash,omitempty"`

	// Storage template configuration
	// +optional
	StorageTemplate *StorageTemplateConfig `json:"storageTemplate,omitempty"`

	// FFmpeg configuration
	// +optional
	FFmpeg *FFmpegConfig `json:"ffmpeg,omitempty"`

	// Job configuration
	// +optional
	Job *JobConfig `json:"job,omitempty"`

	// Library configuration
	// +optional
	Library *LibraryConfig `json:"library,omitempty"`

	// Logging configuration
	// +optional
	Logging *LoggingConfig `json:"logging,omitempty"`

	// MachineLearning configuration
	// +optional
	MachineLearning *MachineLearningConfig `json:"machineLearning,omitempty"`

	// Map configuration
	// +optional
	Map *MapConfig `json:"map,omitempty"`

	// NewVersionCheck configuration
	// +optional
	NewVersionCheck *NewVersionCheckConfig `json:"newVersionCheck,omitempty"`

	// Notifications configuration
	// +optional
	Notifications *NotificationsConfig `json:"notifications,omitempty"`

	// OAuth configuration
	// +optional
	OAuth *OAuthConfig `json:"oauth,omitempty"`

	// PasswordLogin configuration
	// +optional
	PasswordLogin *PasswordLoginConfig `json:"passwordLogin,omitempty"`

	// ReverseGeocoding configuration
	// +optional
	ReverseGeocoding *ReverseGeocodingConfig `json:"reverseGeocoding,omitempty"`

	// Server configuration
	// +optional
	Server *ServerConfig `json:"server,omitempty"`

	// Theme configuration
	// +optional
	Theme *ThemeConfig `json:"theme,omitempty"`

	// User configuration
	// +optional
	User *UserConfig `json:"user,omitempty"`
}

// TrashConfig defines trash bin settings
type TrashConfig struct {
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:default=30
	// +optional
	Days *int `json:"days,omitempty"`
}

// StorageTemplateConfig defines storage template settings
type StorageTemplateConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Template *string `json:"template,omitempty"`
}

// FFmpegConfig defines FFmpeg transcoding settings
type FFmpegConfig struct {
	// +optional
	CRF *int `json:"crf,omitempty"`
	// +optional
	Threads *int `json:"threads,omitempty"`
	// +optional
	Preset *string `json:"preset,omitempty"`
	// +optional
	TargetCodec *string `json:"targetVideoCodec,omitempty"`
	// +optional
	AcceptedAudioCodecs []string `json:"acceptedAudioCodecs,omitempty"`
	// +optional
	TargetResolution *string `json:"targetResolution,omitempty"`
	// +optional
	MaxBitrate *string `json:"maxBitrate,omitempty"`
	// +optional
	Bframes *int `json:"bframes,omitempty"`
	// +optional
	Refs *int `json:"refs,omitempty"`
	// +optional
	GOPSize *int `json:"gopSize,omitempty"`
	// +optional
	NPL *int `json:"npl,omitempty"`
	// +optional
	TemporalAQ *bool `json:"temporalAQ,omitempty"`
	// +optional
	CQMode *string `json:"cqMode,omitempty"`
	// +optional
	TwoPass *bool `json:"twoPass,omitempty"`
	// +optional
	PreferredHwDevice *string `json:"preferredHwDevice,omitempty"`
	// +optional
	TranscodePolicy *string `json:"transcode,omitempty"`
	// +optional
	ToneMappingMode *string `json:"tonemap,omitempty"`
	// +optional
	Accel *string `json:"accel,omitempty"`
	// +optional
	AccelDecode *bool `json:"accelDecode,omitempty"`
}

// JobConfig defines job concurrency settings
type JobConfig struct {
	// +optional
	BackgroundTask *JobConcurrency `json:"backgroundTask,omitempty"`
	// +optional
	SmartSearch *JobConcurrency `json:"smartSearch,omitempty"`
	// +optional
	MetadataExtraction *JobConcurrency `json:"metadataExtraction,omitempty"`
	// +optional
	Search *JobConcurrency `json:"search,omitempty"`
	// +optional
	FaceDetection *JobConcurrency `json:"faceDetection,omitempty"`
	// +optional
	Sidecar *JobConcurrency `json:"sidecar,omitempty"`
	// +optional
	Library *JobConcurrency `json:"library,omitempty"`
	// +optional
	Migration *JobConcurrency `json:"migration,omitempty"`
	// +optional
	ThumbnailGeneration *JobConcurrency `json:"thumbnailGeneration,omitempty"`
	// +optional
	VideoConversion *JobConcurrency `json:"videoConversion,omitempty"`
	// +optional
	Notifications *JobConcurrency `json:"notifications,omitempty"`
}

// JobConcurrency defines concurrency for a specific job type
type JobConcurrency struct {
	// +optional
	Concurrency *int `json:"concurrency,omitempty"`
}

// LibraryConfig defines library scanning settings
type LibraryConfig struct {
	// +optional
	Scan *LibraryScanConfig `json:"scan,omitempty"`
	// +optional
	Watch *LibraryWatchConfig `json:"watch,omitempty"`
}

type LibraryScanConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	CronExpression *string `json:"cronExpression,omitempty"`
}

type LibraryWatchConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Level *string `json:"level,omitempty"`
}

// MachineLearningConfig defines ML settings in immich config.
// Follows the structure from https://docs.immich.app/install/config-file/
type MachineLearningConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	URLs []string `json:"urls,omitempty"`
	// +optional
	Clip *ClipConfig `json:"clip,omitempty"`
	// +optional
	DuplicateDetection *DuplicateDetectionConfig `json:"duplicateDetection,omitempty"`
	// +optional
	FacialRecognition *FacialRecognitionConfig `json:"facialRecognition,omitempty"`
}

type ClipConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	ModelName *string `json:"modelName,omitempty"`
}

type DuplicateDetectionConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	MaxDistance *string `json:"maxDistance,omitempty"`
}

type FacialRecognitionConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	ModelName *string `json:"modelName,omitempty"`
	// +optional
	MinScore *string `json:"minScore,omitempty"`
	// +optional
	MaxDistance *string `json:"maxDistance,omitempty"`
	// +optional
	MinFaces *int `json:"minFaces,omitempty"`
}

// MapConfig defines map settings
type MapConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	LightStyle *string `json:"lightStyle,omitempty"`
	// +optional
	DarkStyle *string `json:"darkStyle,omitempty"`
}

// NewVersionCheckConfig defines version check settings
type NewVersionCheckConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// NotificationsConfig defines notification settings
type NotificationsConfig struct {
	// +optional
	SMTP *SMTPConfig `json:"smtp,omitempty"`
}

type SMTPConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	From *string `json:"from,omitempty"`
	// +optional
	ReplyTo *string `json:"replyTo,omitempty"`
	// +optional
	Transport *SMTPTransportConfig `json:"transport,omitempty"`
}

type SMTPTransportConfig struct {
	// +optional
	Host *string `json:"host,omitempty"`
	// +optional
	Port *int `json:"port,omitempty"`
	// Username for SMTP authentication
	// +optional
	Username *string `json:"username,omitempty"`
	// Reference to a secret containing the SMTP password
	// +optional
	PasswordSecretRef *SecretKeySelector `json:"passwordSecretRef,omitempty"`
	// +optional
	IgnoreCert *bool `json:"ignoreCert,omitempty"`
}

// OAuthConfig defines OAuth settings
type OAuthConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	IssuerURL *string `json:"issuerUrl,omitempty"`
	// +optional
	ClientID *string `json:"clientId,omitempty"`
	// Reference to a secret containing the OAuth client secret
	// +optional
	ClientSecretRef *SecretKeySelector `json:"clientSecretRef,omitempty"`
	// +optional
	Scope *string `json:"scope,omitempty"`
	// +optional
	StorageLabel *string `json:"storageLabelClaim,omitempty"`
	// +optional
	StorageQuota *string `json:"storageQuotaClaim,omitempty"`
	// +optional
	DefaultStorageQuota *int64 `json:"defaultStorageQuota,omitempty"`
	// +optional
	ButtonText *string `json:"buttonText,omitempty"`
	// +optional
	AutoRegister *bool `json:"autoRegister,omitempty"`
	// +optional
	AutoLaunch *bool `json:"autoLaunch,omitempty"`
	// +optional
	MobileOverrideEnabled *bool `json:"mobileOverrideEnabled,omitempty"`
	// +optional
	MobileRedirectURI *string `json:"mobileRedirectUri,omitempty"`
}

// PasswordLoginConfig defines password login settings
type PasswordLoginConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// ReverseGeocodingConfig defines reverse geocoding settings
type ReverseGeocodingConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// ServerConfig defines server-side settings
type ServerConfig struct {
	// +optional
	ExternalDomain *string `json:"externalDomain,omitempty"`
	// +optional
	LoginPageMessage *string `json:"loginPageMessage,omitempty"`
}

// ThemeConfig defines theme settings
type ThemeConfig struct {
	// +optional
	CustomCSS *string `json:"customCss,omitempty"`
}

// UserConfig defines user settings
type UserConfig struct {
	// +optional
	DeleteDelay *int `json:"deleteDelay,omitempty"`
}

// MetricsSpec defines Prometheus metrics configuration.
type MetricsSpec struct {
	// Enable Prometheus metrics and ServiceMonitor creation
	// +kubebuilder:default=false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// PersistenceSpec defines persistence configuration.
type PersistenceSpec struct {
	// Library persistence configuration for photo storage
	// +optional
	Library *LibraryPersistenceSpec `json:"library,omitempty"`
}

// LibraryPersistenceSpec defines library persistence configuration.
// Either use an existing PVC (existingClaim) or let the operator create one (size).
type LibraryPersistenceSpec struct {
	// ExistingClaim is the name of an existing PVC to use for library storage.
	// If set, the operator will use this PVC instead of creating a new one.
	// +optional
	ExistingClaim *string `json:"existingClaim,omitempty"`

	// Size of the PVC to create for library storage.
	// Only used if existingClaim is not set.
	// +kubebuilder:default="10Gi"
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// StorageClass for the PVC. If not set, the default storage class is used.
	// Only used if existingClaim is not set.
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// AccessModes for the PVC.
	// Only used if existingClaim is not set.
	// +kubebuilder:default={"ReadWriteOnce"}
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
}

// ServerSpec defines the server component configuration.
type ServerSpec struct {
	// Enable the server component
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the full image reference (e.g., "ghcr.io/immich-app/immich-server:v1.125.7")
	// If not set, defaults to RELATED_IMAGE_immich environment variable
	// +optional
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy overrides the default pull policy for this component
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Number of replicas
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resource requirements
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Additional environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Additional environment variables from sources
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Ingress configuration (for standard Kubernetes)
	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`

	// Route configuration (for OpenShift)
	// Use this instead of Ingress when running on OpenShift
	// +optional
	Route *RouteSpec `json:"route,omitempty"`

	// Pod annotations
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Pod labels
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// SecurityContext for the pod
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// SecurityContext for the container
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

// MachineLearningSpec defines the machine learning component configuration.
// When enabled=true (default), the operator deploys an ML Deployment.
// When enabled=false, ML is disabled unless an external URL is provided.
// ML is optional - Immich works without it but lacks smart search, face detection, etc.
type MachineLearningSpec struct {
	// Enable the built-in machine learning component
	// Set to false to disable ML or use an external service
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the full image reference (e.g., "ghcr.io/immich-app/immich-machine-learning:v1.125.7")
	// If not set, defaults to RELATED_IMAGE_machineLearning environment variable
	// +optional
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy overrides the default pull policy for this component
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Number of replicas
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resource requirements
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Additional environment variables
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Additional environment variables from sources
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Persistence configuration for ML cache
	// +optional
	Persistence *MachineLearningPersistenceSpec `json:"persistence,omitempty"`

	// Pod annotations
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Pod labels
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// SecurityContext for the pod
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// SecurityContext for the container
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// --- External ML service configuration (used when enabled=false) ---

	// URL of the external ML service (optional, used when enabled=false)
	// If not set when enabled=false, Immich runs without ML features
	// Example: "http://external-ml-service:3003"
	// +optional
	URL *string `json:"url,omitempty"`
}

// MachineLearningPersistenceSpec defines ML cache persistence.
type MachineLearningPersistenceSpec struct {
	// Enable persistence for ML cache
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Size of the cache PVC
	// +kubebuilder:default="10Gi"
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// StorageClass for the cache PVC
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Access modes for the cache PVC
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Use an existing PVC instead of creating one
	// +optional
	ExistingClaim *string `json:"existingClaim,omitempty"`
}

// ValkeySpec defines the Valkey (Redis) component configuration.
// When enabled=true (default), the operator deploys a Valkey StatefulSet.
// When enabled=false, you must provide external Redis connection details.
type ValkeySpec struct {
	// Enable the built-in Valkey component
	// Set to false if using an external Redis/Valkey instance
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the full image reference (e.g., "docker.io/valkey/valkey:9-alpine")
	// If not set, defaults to RELATED_IMAGE_valkey environment variable
	// +optional
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy overrides the default pull policy for this component
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Resource requirements
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence configuration for Valkey data
	// +optional
	Persistence *ValkeyPersistenceSpec `json:"persistence,omitempty"`

	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Pod annotations
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Pod labels
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// SecurityContext for the pod
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// SecurityContext for the container
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// --- External Redis/Valkey configuration (used when enabled=false) ---

	// Hostname of the external Redis/Valkey server (required when enabled=false)
	// +optional
	Host *string `json:"host,omitempty"`

	// Port of the external Redis/Valkey server
	// +kubebuilder:default=6379
	// +optional
	Port *int32 `json:"port,omitempty"`

	// Database index to use (0-15)
	// +kubebuilder:default=0
	// +optional
	DbIndex *int32 `json:"dbIndex,omitempty"`

	// Reference to a secret containing the Redis password
	// +optional
	PasswordSecretRef *SecretKeySelector `json:"passwordSecretRef,omitempty"`
}

// PostgresPersistenceSpec defines PostgreSQL persistence.
type PostgresPersistenceSpec struct {
	// Size of the data PVC
	// +kubebuilder:default="10Gi"
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// StorageClass for the data PVC
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Access modes for the data PVC
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Use an existing PVC instead of creating one
	// +optional
	ExistingClaim *string `json:"existingClaim,omitempty"`
}

// ValkeyPersistenceSpec defines Valkey persistence.
type ValkeyPersistenceSpec struct {
	// Enable persistence for Valkey data
	// +kubebuilder:default=false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Size of the data PVC
	// +kubebuilder:default="10Gi"
	// +optional
	Size *resource.Quantity `json:"size,omitempty"`

	// StorageClass for the data PVC
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Access modes for the data PVC
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Use an existing PVC instead of creating one
	// +optional
	ExistingClaim *string `json:"existingClaim,omitempty"`
}

// PostgresSpec defines PostgreSQL database configuration.
// When enabled=true (default), the operator deploys a PostgreSQL StatefulSet.
// When enabled=false, you must provide external database connection details.
type PostgresSpec struct {
	// Enable the built-in PostgreSQL deployment
	// Set to false if using an external PostgreSQL instance
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image is the full image reference for the PostgreSQL container
	// Must include the pgvecto.rs extension for Immich to work
	// If not set, defaults to RELATED_IMAGE_postgres environment variable
	// +optional
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy overrides the default pull policy for this component
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Resource requirements for the PostgreSQL container
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence configuration for PostgreSQL data
	// +optional
	Persistence *PostgresPersistenceSpec `json:"persistence,omitempty"`

	// Node selector
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Pod annotations
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Pod labels
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// SecurityContext for the pod
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// SecurityContext for the container
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// --- External PostgreSQL configuration (used when enabled=false) ---

	// Hostname of the external PostgreSQL server (required when enabled=false)
	// +optional
	Host *string `json:"host,omitempty"`

	// Port of the PostgreSQL server
	// +kubebuilder:default=5432
	// +optional
	Port *int32 `json:"port,omitempty"`

	// Database name
	// +kubebuilder:default="immich"
	// +optional
	Database *string `json:"database,omitempty"`

	// Username for database connection
	// +kubebuilder:default="immich"
	// +optional
	Username *string `json:"username,omitempty"`

	// Reference to a secret containing the password
	// Required if enabled is false and URLSecretRef is not set
	// +optional
	PasswordSecretRef *SecretKeySelector `json:"passwordSecretRef,omitempty"`

	// Reference to a secret containing the full DATABASE_URL
	// If set, overrides host/port/database/username/password
	// +optional
	URLSecretRef *SecretKeySelector `json:"urlSecretRef,omitempty"`
}

// SecretKeySelector selects a key from a Secret.
type SecretKeySelector struct {
	// Name of the secret
	Name string `json:"name"`
	// Key in the secret
	Key string `json:"key"`
}

// IngressSpec defines ingress configuration.
type IngressSpec struct {
	// Enable ingress
	// +kubebuilder:default=false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Ingress class name
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty"`

	// Annotations for the ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Hosts configuration
	// +optional
	Hosts []IngressHost `json:"hosts,omitempty"`

	// TLS configuration
	// +optional
	TLS []IngressTLS `json:"tls,omitempty"`
}

// IngressHost defines a host for the ingress.
type IngressHost struct {
	// Host name
	// +optional
	Host *string `json:"host,omitempty"`

	// Paths for this host
	// +optional
	Paths []IngressPath `json:"paths,omitempty"`
}

// IngressPath defines a path for the ingress.
type IngressPath struct {
	// Path
	// +kubebuilder:default="/"
	// +optional
	Path *string `json:"path,omitempty"`

	// Path type
	// +kubebuilder:default="Prefix"
	// +optional
	PathType *string `json:"pathType,omitempty"`
}

// IngressTLS defines TLS configuration for the ingress.
type IngressTLS struct {
	// Hosts covered by the TLS certificate
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// Secret name containing the TLS certificate
	// +optional
	SecretName *string `json:"secretName,omitempty"`
}

// RouteSpec defines OpenShift Route configuration.
// On OpenShift clusters, Routes are created by default unless explicitly disabled.
// On non-OpenShift clusters, Routes are not created unless an Ingress is configured.
type RouteSpec struct {
	// Enable OpenShift Route. If not set, auto-detects based on cluster capabilities.
	// Set to false to explicitly disable Route creation on OpenShift.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Host is the hostname for the route (optional, OpenShift will generate one if not set)
	// +optional
	Host *string `json:"host,omitempty"`

	// Path is the path for the route
	// +kubebuilder:default="/"
	// +optional
	Path *string `json:"path,omitempty"`

	// WildcardPolicy defines the wildcard policy for the route
	// +kubebuilder:validation:Enum=None;Subdomain
	// +kubebuilder:default="None"
	// +optional
	WildcardPolicy *string `json:"wildcardPolicy,omitempty"`

	// Annotations for the route
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels for the route
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// TLS configuration for the route
	// +optional
	TLS *RouteTLSConfig `json:"tls,omitempty"`
}

// RouteTLSConfig defines TLS configuration for the OpenShift Route.
type RouteTLSConfig struct {
	// Termination indicates termination type.
	// +kubebuilder:validation:Enum=edge;passthrough;reencrypt
	// +kubebuilder:default="edge"
	// +optional
	Termination *string `json:"termination,omitempty"`

	// InsecureEdgeTerminationPolicy indicates the desired behavior for
	// insecure connections to a route.
	// +kubebuilder:validation:Enum=Allow;Disable;Redirect;None
	// +kubebuilder:default="Redirect"
	// +optional
	InsecureEdgeTerminationPolicy *string `json:"insecureEdgeTerminationPolicy,omitempty"`

	// Certificate is the PEM-encoded certificate (optional, uses default certificate if not set)
	// +optional
	Certificate *string `json:"certificate,omitempty"`

	// Key is the PEM-encoded private key (optional)
	// +optional
	Key *string `json:"key,omitempty"`

	// CACertificate is the PEM-encoded CA certificate (optional)
	// +optional
	CACertificate *string `json:"caCertificate,omitempty"`

	// DestinationCACertificate is the PEM-encoded CA certificate for the backend (used with reencrypt)
	// +optional
	DestinationCACertificate *string `json:"destinationCACertificate,omitempty"`
}

// ImmichStatus defines the observed state of Immich.
type ImmichStatus struct {
	// Conditions represent the latest available observations of the Immich's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Ready indicates if all components are ready
	// +optional
	Ready bool `json:"ready,omitempty"`

	// ServerReady indicates if the server component is ready
	// +optional
	ServerReady bool `json:"serverReady,omitempty"`

	// MachineLearningReady indicates if the machine learning component is ready
	// +optional
	MachineLearningReady bool `json:"machineLearningReady,omitempty"`

	// ValkeyReady indicates if the Valkey component is ready
	// +optional
	ValkeyReady bool `json:"valkeyReady,omitempty"`

	// PostgresReady indicates if the PostgreSQL component is ready
	// +optional
	PostgresReady bool `json:"postgresReady,omitempty"`

	// ObservedGeneration is the last observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// URL is the URL to access Immich (from Route or Ingress)
	// +optional
	URL string `json:"url,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Whether all components are ready"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url",description="URL to access Immich"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Immich is the Schema for the immiches API.
type Immich struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImmichSpec   `json:"spec,omitempty"`
	Status ImmichStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImmichList contains a list of Immich.
type ImmichList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Immich `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Immich{}, &ImmichList{})
}

// Helper methods

// IsServerEnabled returns true if the server component is enabled
func (i *Immich) IsServerEnabled() bool {
	if i.Spec.Server == nil || i.Spec.Server.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.Server.Enabled
}

// IsMachineLearningEnabled returns true if the ML component is enabled
func (i *Immich) IsMachineLearningEnabled() bool {
	if i.Spec.MachineLearning == nil || i.Spec.MachineLearning.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.MachineLearning.Enabled
}

// IsValkeyEnabled returns true if the Valkey component is enabled
func (i *Immich) IsValkeyEnabled() bool {
	if i.Spec.Valkey == nil || i.Spec.Valkey.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.Valkey.Enabled
}

// GetServerImage returns the full server image reference
// Priority order:
// 1. spec.server.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_immich environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetServerImage() string {
	// User-specified image takes precedence
	if i.Spec.Server != nil && i.Spec.Server.Image != nil && *i.Spec.Server.Image != "" {
		return *i.Spec.Server.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageImmich)
}

// GetMachineLearningImage returns the full ML image reference
// Priority order:
// 1. spec.machineLearning.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_machineLearning environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetMachineLearningImage() string {
	// User-specified image takes precedence
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Image != nil && *i.Spec.MachineLearning.Image != "" {
		return *i.Spec.MachineLearning.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageMachineLearning)
}

// GetValkeyImage returns the full Valkey image reference
// Priority order:
// 1. spec.valkey.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_valkey environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetValkeyImage() string {
	// User-specified image takes precedence
	if i.Spec.Valkey != nil && i.Spec.Valkey.Image != nil && *i.Spec.Valkey.Image != "" {
		return *i.Spec.Valkey.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageValkey)
}

// GetLibraryPVCName returns the name of the PVC to use for the photo library.
// Returns the existingClaim if set, otherwise generates a name based on the Immich resource name.
func (i *Immich) GetLibraryPVCName() string {
	if i.Spec.Immich != nil && i.Spec.Immich.Persistence != nil && i.Spec.Immich.Persistence.Library != nil {
		if i.Spec.Immich.Persistence.Library.ExistingClaim != nil && *i.Spec.Immich.Persistence.Library.ExistingClaim != "" {
			return *i.Spec.Immich.Persistence.Library.ExistingClaim
		}
	}
	return i.Name + "-library"
}

// ShouldCreateLibraryPVC returns true if the operator should create a PVC for the library.
// This is true when existingClaim is not set (a default size will be used if not specified).
func (i *Immich) ShouldCreateLibraryPVC() bool {
	if i.Spec.Immich != nil && i.Spec.Immich.Persistence != nil && i.Spec.Immich.Persistence.Library != nil {
		return i.Spec.Immich.Persistence.Library.ExistingClaim == nil || *i.Spec.Immich.Persistence.Library.ExistingClaim == ""
	}
	return true // default to creating a PVC
}

// GetLibrarySize returns the size for the library PVC.
// Defaults to 10Gi if not specified.
func (i *Immich) GetLibrarySize() resource.Quantity {
	if i.Spec.Immich != nil && i.Spec.Immich.Persistence != nil && i.Spec.Immich.Persistence.Library != nil {
		if i.Spec.Immich.Persistence.Library.Size != nil && !i.Spec.Immich.Persistence.Library.Size.IsZero() {
			return *i.Spec.Immich.Persistence.Library.Size
		}
	}
	return resource.MustParse("10Gi")
}

// GetLibraryAccessModes returns the access modes for the library PVC.
// Defaults to ReadWriteOnce if not specified.
func (i *Immich) GetLibraryAccessModes() []corev1.PersistentVolumeAccessMode {
	if i.Spec.Immich != nil && i.Spec.Immich.Persistence != nil && i.Spec.Immich.Persistence.Library != nil {
		if len(i.Spec.Immich.Persistence.Library.AccessModes) > 0 {
			return i.Spec.Immich.Persistence.Library.AccessModes
		}
	}
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
}

// GetLibraryStorageClass returns the storage class for the library PVC.
func (i *Immich) GetLibraryStorageClass() *string {
	if i.Spec.Immich != nil && i.Spec.Immich.Persistence != nil && i.Spec.Immich.Persistence.Library != nil {
		return i.Spec.Immich.Persistence.Library.StorageClass
	}
	return nil
}

// IsPostgresEnabled returns true if the built-in PostgreSQL is enabled
func (i *Immich) IsPostgresEnabled() bool {
	if i.Spec.Postgres == nil || i.Spec.Postgres.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.Postgres.Enabled
}

// GetPostgresImage returns the full PostgreSQL image reference
// Priority order:
// 1. spec.postgres.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_postgres environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetPostgresImage() string {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Image != nil && *i.Spec.Postgres.Image != "" {
		return *i.Spec.Postgres.Image
	}
	return os.Getenv(EnvRelatedImagePostgres)
}

// GetImmichInitContainerImage returns the image to use for Immich init containers.
// Falls back to RELATED_IMAGE_immich_initContainer environment variable.
func GetImmichInitContainerImage() string {
	return os.Getenv(EnvRelatedImageImmichInitContainer)
}

// GetPostgresPVCName returns the name of the PVC for PostgreSQL data.
// When using VolumeClaimTemplates, the PVC is named: <volumeClaimTemplate.name>-<statefulset.name>-<ordinal>
func (i *Immich) GetPostgresPVCName() string {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Persistence != nil {
		if i.Spec.Postgres.Persistence.ExistingClaim != nil && *i.Spec.Postgres.Persistence.ExistingClaim != "" {
			return *i.Spec.Postgres.Persistence.ExistingClaim
		}
	}
	// VolumeClaimTemplate name is "data", StatefulSet name is "<immich.name>-postgres", ordinal is 0
	return "data-" + i.Name + "-postgres-0"
}

// GetPostgresHost returns the hostname to connect to PostgreSQL.
// If built-in is enabled, returns the service name. Otherwise returns the external host.
func (i *Immich) GetPostgresHost() string {
	if i.IsPostgresEnabled() {
		return i.Name + "-postgres"
	}
	if i.Spec.Postgres != nil && i.Spec.Postgres.Host != nil {
		return *i.Spec.Postgres.Host
	}
	return ""
}

// GetPostgresPort returns the port for PostgreSQL connection.
func (i *Immich) GetPostgresPort() int32 {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Port != nil && *i.Spec.Postgres.Port != 0 {
		return *i.Spec.Postgres.Port
	}
	return 5432
}

// GetPostgresDatabase returns the database name.
func (i *Immich) GetPostgresDatabase() string {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Database != nil && *i.Spec.Postgres.Database != "" {
		return *i.Spec.Postgres.Database
	}
	return "immich"
}

// GetPostgresUsername returns the username for PostgreSQL.
func (i *Immich) GetPostgresUsername() string {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Username != nil && *i.Spec.Postgres.Username != "" {
		return *i.Spec.Postgres.Username
	}
	return "immich"
}

// GetValkeyHost returns the hostname to connect to Valkey/Redis.
// If built-in is enabled, returns the service name. Otherwise returns the external host.
func (i *Immich) GetValkeyHost() string {
	if i.IsValkeyEnabled() {
		return i.Name + "-valkey"
	}
	if i.Spec.Valkey != nil && i.Spec.Valkey.Host != nil {
		return *i.Spec.Valkey.Host
	}
	return ""
}

// GetValkeyPort returns the port for Valkey/Redis connection.
func (i *Immich) GetValkeyPort() int32 {
	if i.Spec.Valkey != nil && i.Spec.Valkey.Port != nil && *i.Spec.Valkey.Port != 0 {
		return *i.Spec.Valkey.Port
	}
	return 6379
}

// GetMachineLearningURL returns the URL for the machine learning service.
// If built-in is enabled, returns the internal service URL. Otherwise returns the external URL.
func (i *Immich) GetMachineLearningURL() string {
	if i.IsMachineLearningEnabled() {
		return "http://" + i.Name + "-machine-learning:3003"
	}
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.URL != nil {
		return *i.Spec.MachineLearning.URL
	}
	return ""
}

// IsIngressEnabled returns true if ingress is enabled for the server
func (i *Immich) IsIngressEnabled() bool {
	if i.Spec.Server == nil || i.Spec.Server.Ingress == nil || i.Spec.Server.Ingress.Enabled == nil {
		return false // default to disabled
	}
	return *i.Spec.Server.Ingress.Enabled
}

// IsRouteEnabled returns true if OpenShift Route is explicitly enabled for the server
func (i *Immich) IsRouteEnabled() bool {
	if i.Spec.Server == nil || i.Spec.Server.Route == nil || i.Spec.Server.Route.Enabled == nil {
		return false
	}
	return *i.Spec.Server.Route.Enabled
}

// IsRouteExplicitlyDisabled returns true if Route is explicitly disabled (set to false)
func (i *Immich) IsRouteExplicitlyDisabled() bool {
	if i.Spec.Server == nil || i.Spec.Server.Route == nil || i.Spec.Server.Route.Enabled == nil {
		return false // not explicitly disabled, just not set
	}
	return !*i.Spec.Server.Route.Enabled
}

// ShouldCreateRoute returns true if a Route should be created
// It creates a Route if:
// - Route API is available AND route is not explicitly disabled
// - OR route is explicitly enabled (even if API check wasn't done)
func (i *Immich) ShouldCreateRoute(routeAPIAvailable bool) bool {
	// If explicitly disabled, don't create
	if i.IsRouteExplicitlyDisabled() {
		return false
	}
	// If explicitly enabled, create
	if i.IsRouteEnabled() {
		return true
	}
	// Auto-detect: create if Route API is available
	return routeAPIAvailable
}

// IsMetricsEnabled returns true if metrics are enabled
func (i *Immich) IsMetricsEnabled() bool {
	if i.Spec.Immich == nil || i.Spec.Immich.Metrics == nil || i.Spec.Immich.Metrics.Enabled == nil {
		return false // default to disabled
	}
	return *i.Spec.Immich.Metrics.Enabled
}

// GetConfigurationKind returns the kind of resource to store configuration in
func (i *Immich) GetConfigurationKind() string {
	if i.Spec.Immich != nil && i.Spec.Immich.ConfigurationKind != nil && *i.Spec.Immich.ConfigurationKind != "" {
		return *i.Spec.Immich.ConfigurationKind
	}
	return "ConfigMap"
}

// GetServerReplicas returns the number of server replicas
func (i *Immich) GetServerReplicas() int32 {
	if i.Spec.Server != nil && i.Spec.Server.Replicas != nil {
		return *i.Spec.Server.Replicas
	}
	return 1
}

// GetMachineLearningReplicas returns the number of ML replicas
func (i *Immich) GetMachineLearningReplicas() int32 {
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Replicas != nil {
		return *i.Spec.MachineLearning.Replicas
	}
	return 1
}

// IsMLPersistenceEnabled returns true if ML cache persistence is enabled
func (i *Immich) IsMLPersistenceEnabled() bool {
	if i.Spec.MachineLearning == nil || i.Spec.MachineLearning.Persistence == nil || i.Spec.MachineLearning.Persistence.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.MachineLearning.Persistence.Enabled
}

// GetMLCachePVCName returns the name of the ML cache PVC
func (i *Immich) GetMLCachePVCName() string {
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Persistence != nil {
		if i.Spec.MachineLearning.Persistence.ExistingClaim != nil && *i.Spec.MachineLearning.Persistence.ExistingClaim != "" {
			return *i.Spec.MachineLearning.Persistence.ExistingClaim
		}
	}
	return i.Name + "-ml-cache"
}

// GetMLCacheSize returns the size for the ML cache PVC
func (i *Immich) GetMLCacheSize() resource.Quantity {
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Persistence != nil {
		if i.Spec.MachineLearning.Persistence.Size != nil && !i.Spec.MachineLearning.Persistence.Size.IsZero() {
			return *i.Spec.MachineLearning.Persistence.Size
		}
	}
	return resource.MustParse("10Gi")
}

// GetMLCacheAccessModes returns the access modes for the ML cache PVC
func (i *Immich) GetMLCacheAccessModes() []corev1.PersistentVolumeAccessMode {
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Persistence != nil {
		if len(i.Spec.MachineLearning.Persistence.AccessModes) > 0 {
			return i.Spec.MachineLearning.Persistence.AccessModes
		}
	}
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
}

// GetMLCacheStorageClass returns the storage class for the ML cache PVC
func (i *Immich) GetMLCacheStorageClass() *string {
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Persistence != nil {
		return i.Spec.MachineLearning.Persistence.StorageClass
	}
	return nil
}

// GetPostgresSize returns the size for the PostgreSQL PVC
func (i *Immich) GetPostgresSize() resource.Quantity {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Persistence != nil {
		if i.Spec.Postgres.Persistence.Size != nil && !i.Spec.Postgres.Persistence.Size.IsZero() {
			return *i.Spec.Postgres.Persistence.Size
		}
	}
	return resource.MustParse("10Gi")
}

// GetPostgresAccessModes returns the access modes for the PostgreSQL PVC
func (i *Immich) GetPostgresAccessModes() []corev1.PersistentVolumeAccessMode {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Persistence != nil {
		if len(i.Spec.Postgres.Persistence.AccessModes) > 0 {
			return i.Spec.Postgres.Persistence.AccessModes
		}
	}
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
}

// GetPostgresStorageClass returns the storage class for the PostgreSQL PVC
func (i *Immich) GetPostgresStorageClass() *string {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Persistence != nil {
		return i.Spec.Postgres.Persistence.StorageClass
	}
	return nil
}

// IsValkeyPersistenceEnabled returns true if Valkey persistence is enabled
func (i *Immich) IsValkeyPersistenceEnabled() bool {
	if i.Spec.Valkey == nil || i.Spec.Valkey.Persistence == nil || i.Spec.Valkey.Persistence.Enabled == nil {
		return false // default to disabled
	}
	return *i.Spec.Valkey.Persistence.Enabled
}

// GetValkeyPVCName returns the name of the Valkey PVC
func (i *Immich) GetValkeyPVCName() string {
	if i.Spec.Valkey != nil && i.Spec.Valkey.Persistence != nil {
		if i.Spec.Valkey.Persistence.ExistingClaim != nil && *i.Spec.Valkey.Persistence.ExistingClaim != "" {
			return *i.Spec.Valkey.Persistence.ExistingClaim
		}
	}
	return i.Name + "-valkey-data"
}

// GetValkeySize returns the size for the Valkey PVC
func (i *Immich) GetValkeySize() resource.Quantity {
	if i.Spec.Valkey != nil && i.Spec.Valkey.Persistence != nil {
		if i.Spec.Valkey.Persistence.Size != nil && !i.Spec.Valkey.Persistence.Size.IsZero() {
			return *i.Spec.Valkey.Persistence.Size
		}
	}
	return resource.MustParse("10Gi")
}

// GetValkeyAccessModes returns the access modes for the Valkey PVC
func (i *Immich) GetValkeyAccessModes() []corev1.PersistentVolumeAccessMode {
	if i.Spec.Valkey != nil && i.Spec.Valkey.Persistence != nil {
		if len(i.Spec.Valkey.Persistence.AccessModes) > 0 {
			return i.Spec.Valkey.Persistence.AccessModes
		}
	}
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
}

// GetValkeyStorageClass returns the storage class for the Valkey PVC
func (i *Immich) GetValkeyStorageClass() *string {
	if i.Spec.Valkey != nil && i.Spec.Valkey.Persistence != nil {
		return i.Spec.Valkey.Persistence.StorageClass
	}
	return nil
}

// GetValkeyDbIndex returns the database index for Valkey
func (i *Immich) GetValkeyDbIndex() int32 {
	if i.Spec.Valkey != nil && i.Spec.Valkey.DbIndex != nil {
		return *i.Spec.Valkey.DbIndex
	}
	return 0
}

// ShouldCreateMLCachePVC returns true if the operator should create a PVC for ML cache
func (i *Immich) ShouldCreateMLCachePVC() bool {
	if !i.IsMLPersistenceEnabled() {
		return false
	}
	if i.Spec.MachineLearning != nil && i.Spec.MachineLearning.Persistence != nil {
		return i.Spec.MachineLearning.Persistence.ExistingClaim == nil || *i.Spec.MachineLearning.Persistence.ExistingClaim == ""
	}
	return true
}

// ShouldCreateValkeyPVC returns true if the operator should create a PVC for Valkey
func (i *Immich) ShouldCreateValkeyPVC() bool {
	if !i.IsValkeyPersistenceEnabled() {
		return false
	}
	if i.Spec.Valkey != nil && i.Spec.Valkey.Persistence != nil {
		return i.Spec.Valkey.Persistence.ExistingClaim == nil || *i.Spec.Valkey.Persistence.ExistingClaim == ""
	}
	return true
}

// ShouldCreatePostgresPVC returns true if the operator should create a PVC for PostgreSQL
func (i *Immich) ShouldCreatePostgresPVC() bool {
	if i.Spec.Postgres != nil && i.Spec.Postgres.Persistence != nil {
		return i.Spec.Postgres.Persistence.ExistingClaim == nil || *i.Spec.Postgres.Persistence.ExistingClaim == ""
	}
	return true
}
