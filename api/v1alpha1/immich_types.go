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
	EnvRelatedImageImmich          = "RELATED_IMAGE_immich"
	EnvRelatedImageMachineLearning = "RELATED_IMAGE_machineLearning"
	EnvRelatedImageValkey          = "RELATED_IMAGE_valkey"
)

// ImmichSpec defines the desired state of Immich.
type ImmichSpec struct {
	// ImagePullSecrets are the secrets used to pull images from private registries
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Immich shared configuration
	// +optional
	Immich ImmichConfig `json:"immich,omitempty"`

	// Server component configuration
	// +optional
	Server ServerSpec `json:"server,omitempty"`

	// MachineLearning component configuration
	// +optional
	MachineLearning MachineLearningSpec `json:"machineLearning,omitempty"`

	// Valkey (Redis) component configuration
	// +optional
	Valkey ValkeySpec `json:"valkey,omitempty"`

	// PostgreSQL database configuration
	// +optional
	Postgres PostgresSpec `json:"postgres,omitempty"`
}

// ImmichConfig defines shared Immich configuration.
type ImmichConfig struct {
	// Metrics configuration
	// +optional
	Metrics MetricsSpec `json:"metrics,omitempty"`

	// Persistence configuration for photo library
	// +optional
	Persistence PersistenceSpec `json:"persistence,omitempty"`

	// Configuration is immich-config.yaml converted to raw YAML
	// ref: https://immich.app/docs/install/config-file/
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// ConfigurationKind sets the resource Kind to store configuration in.
	// Must be either ConfigMap or Secret.
	// +kubebuilder:validation:Enum=ConfigMap;Secret
	// +kubebuilder:default="ConfigMap"
	// +optional
	ConfigurationKind string `json:"configurationKind,omitempty"`
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
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default=30
	Days int `json:"days,omitempty"`
}

// StorageTemplateConfig defines storage template settings
type StorageTemplateConfig struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Template string `json:"template,omitempty"`
}

// FFmpegConfig defines FFmpeg transcoding settings
type FFmpegConfig struct {
	CRF                 int      `json:"crf,omitempty"`
	Threads             int      `json:"threads,omitempty"`
	Preset              string   `json:"preset,omitempty"`
	TargetCodec         string   `json:"targetVideoCodec,omitempty"`
	AcceptedAudioCodecs []string `json:"acceptedAudioCodecs,omitempty"`
	TargetResolution    string   `json:"targetResolution,omitempty"`
	MaxBitrate          string   `json:"maxBitrate,omitempty"`
	Bframes             int      `json:"bframes,omitempty"`
	Refs                int      `json:"refs,omitempty"`
	GOPSize             int      `json:"gopSize,omitempty"`
	NPL                 int      `json:"npl,omitempty"`
	TemporalAQ          bool     `json:"temporalAQ,omitempty"`
	CQMode              string   `json:"cqMode,omitempty"`
	TwoPass             bool     `json:"twoPass,omitempty"`
	PreferredHwDevice   string   `json:"preferredHwDevice,omitempty"`
	TranscodePolicy     string   `json:"transcode,omitempty"`
	ToneMappingMode     string   `json:"tonemap,omitempty"`
	Accel               string   `json:"accel,omitempty"`
	AccelDecode         bool     `json:"accelDecode,omitempty"`
}

// JobConfig defines job concurrency settings
type JobConfig struct {
	BackgroundTask      *JobConcurrency `json:"backgroundTask,omitempty"`
	SmartSearch         *JobConcurrency `json:"smartSearch,omitempty"`
	MetadataExtraction  *JobConcurrency `json:"metadataExtraction,omitempty"`
	Search              *JobConcurrency `json:"search,omitempty"`
	FaceDetection       *JobConcurrency `json:"faceDetection,omitempty"`
	Sidecar             *JobConcurrency `json:"sidecar,omitempty"`
	Library             *JobConcurrency `json:"library,omitempty"`
	Migration           *JobConcurrency `json:"migration,omitempty"`
	ThumbnailGeneration *JobConcurrency `json:"thumbnailGeneration,omitempty"`
	VideoConversion     *JobConcurrency `json:"videoConversion,omitempty"`
	Notifications       *JobConcurrency `json:"notifications,omitempty"`
}

// JobConcurrency defines concurrency for a specific job type
type JobConcurrency struct {
	Concurrency int `json:"concurrency,omitempty"`
}

// LibraryConfig defines library scanning settings
type LibraryConfig struct {
	Scan  *LibraryScanConfig  `json:"scan,omitempty"`
	Watch *LibraryWatchConfig `json:"watch,omitempty"`
}

type LibraryScanConfig struct {
	Enabled        bool   `json:"enabled,omitempty"`
	CronExpression string `json:"cronExpression,omitempty"`
}

type LibraryWatchConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Level   string `json:"level,omitempty"`
}

// MachineLearningConfig defines ML settings in immich config
type MachineLearningConfig struct {
	Enabled            bool                      `json:"enabled,omitempty"`
	URL                string                    `json:"url,omitempty"`
	Clip               *ClipConfig               `json:"clip,omitempty"`
	DuplicateDetection *DuplicateDetectionConfig `json:"duplicateDetection,omitempty"`
	FacialRecognition  *FacialRecognitionConfig  `json:"facialRecognition,omitempty"`
}

type ClipConfig struct {
	Enabled   bool   `json:"enabled,omitempty"`
	ModelName string `json:"modelName,omitempty"`
}

type DuplicateDetectionConfig struct {
	Enabled     bool   `json:"enabled,omitempty"`
	MaxDistance string `json:"maxDistance,omitempty"`
}

type FacialRecognitionConfig struct {
	Enabled     bool   `json:"enabled,omitempty"`
	ModelName   string `json:"modelName,omitempty"`
	MinScore    string `json:"minScore,omitempty"`
	MaxDistance string `json:"maxDistance,omitempty"`
	MinFaces    int    `json:"minFaces,omitempty"`
}

// MapConfig defines map settings
type MapConfig struct {
	Enabled    bool   `json:"enabled,omitempty"`
	LightStyle string `json:"lightStyle,omitempty"`
	DarkStyle  string `json:"darkStyle,omitempty"`
}

// NewVersionCheckConfig defines version check settings
type NewVersionCheckConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// NotificationsConfig defines notification settings
type NotificationsConfig struct {
	SMTP *SMTPConfig `json:"smtp,omitempty"`
}

type SMTPConfig struct {
	Enabled   bool                 `json:"enabled,omitempty"`
	From      string               `json:"from,omitempty"`
	ReplyTo   string               `json:"replyTo,omitempty"`
	Transport *SMTPTransportConfig `json:"transport,omitempty"`
}

type SMTPTransportConfig struct {
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	IgnoreCert bool   `json:"ignoreCert,omitempty"`
}

// OAuthConfig defines OAuth settings
type OAuthConfig struct {
	Enabled               bool   `json:"enabled,omitempty"`
	IssuerURL             string `json:"issuerUrl,omitempty"`
	ClientID              string `json:"clientId,omitempty"`
	ClientSecret          string `json:"clientSecret,omitempty"`
	Scope                 string `json:"scope,omitempty"`
	StorageLabel          string `json:"storageLabelClaim,omitempty"`
	StorageQuota          string `json:"storageQuotaClaim,omitempty"`
	DefaultStorageQuota   int64  `json:"defaultStorageQuota,omitempty"`
	ButtonText            string `json:"buttonText,omitempty"`
	AutoRegister          bool   `json:"autoRegister,omitempty"`
	AutoLaunch            bool   `json:"autoLaunch,omitempty"`
	MobileOverrideEnabled bool   `json:"mobileOverrideEnabled,omitempty"`
	MobileRedirectURI     string `json:"mobileRedirectUri,omitempty"`
}

// PasswordLoginConfig defines password login settings
type PasswordLoginConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// ReverseGeocodingConfig defines reverse geocoding settings
type ReverseGeocodingConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

// ServerConfig defines server-side settings
type ServerConfig struct {
	ExternalDomain   string `json:"externalDomain,omitempty"`
	LoginPageMessage string `json:"loginPageMessage,omitempty"`
}

// ThemeConfig defines theme settings
type ThemeConfig struct {
	CustomCSS string `json:"customCss,omitempty"`
}

// UserConfig defines user settings
type UserConfig struct {
	DeleteDelay int `json:"deleteDelay,omitempty"`
}

// MetricsSpec defines Prometheus metrics configuration.
type MetricsSpec struct {
	// Enable Prometheus metrics and ServiceMonitor creation
	// +kubebuilder:default=false
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// PersistenceSpec defines persistence configuration.
type PersistenceSpec struct {
	// Library persistence configuration for photo storage
	// +optional
	Library LibraryPersistenceSpec `json:"library,omitempty"`
}

// LibraryPersistenceSpec defines library persistence configuration.
// Either use an existing PVC (existingClaim) or let the operator create one (size).
type LibraryPersistenceSpec struct {
	// ExistingClaim is the name of an existing PVC to use for library storage.
	// If set, the operator will use this PVC instead of creating a new one.
	// +optional
	ExistingClaim string `json:"existingClaim,omitempty"`

	// Size of the PVC to create for library storage.
	// Only used if existingClaim is not set.
	// +optional
	Size resource.Quantity `json:"size,omitempty"`

	// StorageClass for the PVC. If not set, the default storage class is used.
	// Only used if existingClaim is not set.
	// +optional
	StorageClass string `json:"storageClass,omitempty"`

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

	// Image configuration for the server
	// +optional
	Image ComponentImageSpec `json:"image,omitempty"`

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

	// Ingress configuration
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`

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
type MachineLearningSpec struct {
	// Enable the machine learning component
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image configuration for machine learning
	// +optional
	Image ComponentImageSpec `json:"image,omitempty"`

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
	Persistence MachineLearningPersistenceSpec `json:"persistence,omitempty"`

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

// MachineLearningPersistenceSpec defines ML cache persistence.
type MachineLearningPersistenceSpec struct {
	// Enable persistence for ML cache
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Size of the cache PVC
	// +kubebuilder:default="10Gi"
	// +optional
	Size resource.Quantity `json:"size,omitempty"`

	// StorageClass for the cache PVC
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Access modes for the cache PVC
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Use an existing PVC instead of creating one
	// +optional
	ExistingClaim string `json:"existingClaim,omitempty"`
}

// ValkeySpec defines the Valkey (Redis) component configuration.
type ValkeySpec struct {
	// Enable the built-in Valkey component
	// Set to false if using an external Redis/Valkey instance
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Image configuration for Valkey
	// +optional
	Image ComponentImageSpec `json:"image,omitempty"`

	// Resource requirements
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Persistence configuration for Valkey data
	// +optional
	Persistence ValkeyPersistenceSpec `json:"persistence,omitempty"`

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
}

// ValkeyPersistenceSpec defines Valkey persistence.
type ValkeyPersistenceSpec struct {
	// Enable persistence for Valkey data
	// +kubebuilder:default=false
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Size of the data PVC
	// +kubebuilder:default="1Gi"
	// +optional
	Size resource.Quantity `json:"size,omitempty"`

	// StorageClass for the data PVC
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Access modes for the data PVC
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`

	// Use an existing PVC instead of creating one
	// +optional
	ExistingClaim string `json:"existingClaim,omitempty"`
}

// PostgresSpec defines PostgreSQL database configuration.
type PostgresSpec struct {
	// Hostname of the PostgreSQL server
	Host string `json:"host,omitempty"`

	// Port of the PostgreSQL server
	// +kubebuilder:default=5432
	// +optional
	Port int32 `json:"port,omitempty"`

	// Database name
	// +kubebuilder:default="immich"
	// +optional
	Database string `json:"database,omitempty"`

	// Username for database connection
	// +kubebuilder:default="immich"
	// +optional
	Username string `json:"username,omitempty"`

	// Password for database connection (plain text - prefer PasswordSecretRef)
	// +optional
	Password string `json:"password,omitempty"`

	// Reference to a secret containing the password
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

// ComponentImageSpec defines image configuration for a specific component.
type ComponentImageSpec struct {
	// Image is the full image reference (e.g., "ghcr.io/immich-app/immich-server:v1.125.7")
	// +optional
	Image string `json:"image,omitempty"`

	// PullPolicy overrides the default pull policy for this component
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

// IngressSpec defines ingress configuration.
type IngressSpec struct {
	// Enable ingress
	// +kubebuilder:default=false
	// +optional
	Enabled bool `json:"enabled,omitempty"`

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
	Host string `json:"host,omitempty"`

	// Paths for this host
	// +optional
	Paths []IngressPath `json:"paths,omitempty"`
}

// IngressPath defines a path for the ingress.
type IngressPath struct {
	// Path
	// +kubebuilder:default="/"
	Path string `json:"path,omitempty"`

	// Path type
	// +kubebuilder:default="Prefix"
	PathType string `json:"pathType,omitempty"`
}

// IngressTLS defines TLS configuration for the ingress.
type IngressTLS struct {
	// Hosts covered by the TLS certificate
	Hosts []string `json:"hosts,omitempty"`

	// Secret name containing the TLS certificate
	SecretName string `json:"secretName,omitempty"`
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

	// ObservedGeneration is the last observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Whether all components are ready"
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
	if i.Spec.Server.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.Server.Enabled
}

// IsMachineLearningEnabled returns true if the ML component is enabled
func (i *Immich) IsMachineLearningEnabled() bool {
	if i.Spec.MachineLearning.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.MachineLearning.Enabled
}

// IsValkeyEnabled returns true if the Valkey component is enabled
func (i *Immich) IsValkeyEnabled() bool {
	if i.Spec.Valkey.Enabled == nil {
		return true // default to enabled
	}
	return *i.Spec.Valkey.Enabled
}

// GetServerImage returns the full server image reference
// Priority order:
// 1. spec.server.image.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_IMMICH environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetServerImage() string {
	// User-specified image takes precedence
	if i.Spec.Server.Image.Image != "" {
		return i.Spec.Server.Image.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageImmich)
}

// GetMachineLearningImage returns the full ML image reference
// Priority order:
// 1. spec.machineLearning.image.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_MACHINE_LEARNING environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetMachineLearningImage() string {
	// User-specified image takes precedence
	if i.Spec.MachineLearning.Image.Image != "" {
		return i.Spec.MachineLearning.Image.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageMachineLearning)
}

// GetValkeyImage returns the full Valkey image reference
// Priority order:
// 1. spec.valkey.image.image (user-specified in CR takes precedence)
// 2. RELATED_IMAGE_VALKEY environment variable (for disconnected environments)
// Returns empty string if neither is set (caller should handle as error)
func (i *Immich) GetValkeyImage() string {
	// User-specified image takes precedence
	if i.Spec.Valkey.Image.Image != "" {
		return i.Spec.Valkey.Image.Image
	}

	// Fall back to environment variable (disconnected/air-gapped support)
	return os.Getenv(EnvRelatedImageValkey)
}

// GetLibraryPVCName returns the name of the PVC to use for the photo library.
// Returns the existingClaim if set, otherwise generates a name based on the Immich resource name.
func (i *Immich) GetLibraryPVCName() string {
	if i.Spec.Immich.Persistence.Library.ExistingClaim != "" {
		return i.Spec.Immich.Persistence.Library.ExistingClaim
	}
	return i.Name + "-library"
}

// ShouldCreateLibraryPVC returns true if the operator should create a PVC for the library.
// This is true when existingClaim is not set but size is configured.
func (i *Immich) ShouldCreateLibraryPVC() bool {
	return i.Spec.Immich.Persistence.Library.ExistingClaim == "" &&
		!i.Spec.Immich.Persistence.Library.Size.IsZero()
}

// GetLibraryAccessModes returns the access modes for the library PVC.
// Defaults to ReadWriteOnce if not specified.
func (i *Immich) GetLibraryAccessModes() []corev1.PersistentVolumeAccessMode {
	if len(i.Spec.Immich.Persistence.Library.AccessModes) > 0 {
		return i.Spec.Immich.Persistence.Library.AccessModes
	}
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
}
