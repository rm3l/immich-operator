/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImmichSpec defines the desired state of Immich
type ImmichSpec struct {

	// Application settings
	Application *Application `json:"application,omitempty"`
}

type Application struct {

	// Domain for public shared links, including http(s)://
	// +optional
	ExternalDomain *string `json:"externalDomain,omitempty"`

	// Domain for public shared links, including http(s)://
	// +optional
	WelcomeMessage *string `json:"welcomeMessage,omitempty"`

	// Cascading Style Sheets allow the design of Immich to be customized.
	// +optional
	CustomCSS *string `json:"customCSS,omitempty"`

	// OAuth Login Settings
	// +optional
	OAuthLogin *OAuthLogin `json:"oauth,omitempty"`

	// Login with email and password
	// +optional
	//+kubebuilder:default=true
	PasswordLoginEnabled *bool `json:"passwordLoginEnabled,omitempty"`

	// Database Backup settings
	// +optional
	DatabaseBackup *DatabaseBackup `json:"databaseBackup,omitempty"`

	// Quality and resolution of generated images
	// +optional
	ImageSettings *ImageSettings `json:"imageSettings,omitempty"`

	// Resolution and encoding information of the video files
	// +optional
	VideoTranscodingSettings *VideoTranscodingSettings `json:"videoTranscodingSettings,omitempty"`

	// Job concurrency settings
	// +optional
	JobConcurrency *JobConcurrency `json:"concurrency,omitempty"`

	// Import faces from image EXIF data and sidecar files
	// +optional
	//+kubebuilder:default=false
	MetadataFaceImportEnabled *bool `json:"metadataFaceImportEnabled,omitempty"`

	// Periodic library scanning
	// +optional
	ExternalLibraryScanning *ExternalLibraryScanningSettings `json:"externalLibraryScanning,omitempty"`

	// Log Settings
	// +optional
	Logging *Logging `json:"logging,omitempty"`

	// Machine Learning Settings
	// +optional
	MachineLearning *MachineLearning `json:"machineLearning,omitempty"`

	// Enable map features with rely on an external tile service (like tiles.immich.cloud).
	// +optional
	//+kubebuilder:default=true
	MapEnabled *bool `json:"mapEnabled,omitempty"`

	// Enable Reverse Geocoding using data from the GeoNames geographical database.
	// +optional
	//+kubebuilder:default=true
	MapReverseGeocodingEnabled *bool `json:"mapEnabled,omitempty"`

	// Map Light style.
	// +optional
	//+kubebuilder:default="https://tiles.immich.cloud/v1/style/light.json"
	MapLightStyleURL *string `json:"mapLightStyleURL,omitempty"`

	// Map Dark style.
	// +optional
	//+kubebuilder:default="https://tiles.immich.cloud/v1/style/dark.json"
	MapDarkStyleURL *string `json:"mapDarkStyleURL,omitempty"`

	// Settings for sending email notifications
	// +optional
	EmailNotifications *EmailNotifications `json:"email,omitempty"`

	// This manages the folder structure and filename of the upload asset.
	//+kubebuilder:default=true
	// +optional
	StorageTemplateEnabled *bool `json:"storageTemplateEnabled,omitempty"`

	// Enables hash verification, don't disable this unless you're certain of the implications
	// +kubebuilder:default=true
	// +optional
	StorageTemplateHashVerificationEnabled *bool `json:"storageTemplateHashVerificationEnabled,omitempty"`

	// Enables hash verification, don't disable this unless you're certain of the implications
	// +kubebuilder:default="{{y}}/{{y}}-{{MM}}-{{dd}}/{{filename}}"
	// +optional
	StorageTemplateTemplate *string `json:"storageTemplateTemplate,omitempty"`

	// Experimental settings
	// +optional
	ExperimentalSettings *ExperimentalSettings `json:"experimentalSettings,omitempty"`

	// Enable trash
	// +kubebuilder:default=true
	// +optional
	TrashEnabled *bool `json:"trashEnabled,omitempty"`

	// Number of days to keep the assets in trash before permanently removing them
	// +kubebuilder:default=30
	// +optional
	TrashTTLDays *int32 `json:"trashTTLDays,omitempty"`

	// Number of days after removal to permanently delete a user's account and assets.
	// The user deletion job runs at midnight to check for users that are ready for deletion.
	// Changes to this setting will be evaluated at the next execution.
	// +kubebuilder:default=7
	// +optional
	UserDeleteTTLDays *bool `json:"userDeleteTTLDays,omitempty"`

	// Check for new versions of Immich. The version check feature relies on periodic communication with github.com
	// +kubebuilder:default=true
	// +optional
	VersionCheckEnabled *bool `json:"versionCheckEnabled,omitempty"`
}

type EmailNotifications struct {
	// Enable email notifications.
	// +optional
	//+kubebuilder:default=false
	Enabled *bool `json:"enabled,omitempty"`

	// Host of the email server (e.g. smtp.immich.app)
	// +kubebuilder:example="smtp.immich.app"
	// +optional
	ServerHostSecretKeyRef *SecretKeyRef `json:"serverHostSecretKeyRef,omitempty"`

	// Port of the email server (e.g 25, 465, or 587)
	// +kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=65535
	// +kubebuilder:example=25
	// +kubebuilder:example=465
	// +kubebuilder:example=587
	// +optional
	ServerPort *int32 `json:"serverPort,omitempty"`

	// Ignore TLS certificate validation errors (not recommended)
	// +optional
	//+kubebuilder:default=false
	SkipCertificateChecks *bool `json:"skipCertificateChecks,omitempty"`

	// Username to use when authenticating with the email server
	// +optional
	UsernameSecretKeyRef *SecretKeyRef `json:"usernameSecretKeyRef,omitempty"`

	// Password to use when authenticating with the email server
	// +optional
	PasswordSecretKeyRef *SecretKeyRef `json:"passwordSecretKeyRef,omitempty"`

	// Secret containing the sender email address.
	// The secret could contain something like: "Immich Photo Server <noreply@example.com>"
	// +optional
	FromAddressSecretKeyRef *SecretKeyRef `json:"fromAddressSecretKeyRef,omitempty"`
}

type MachineLearning struct {
	// Enable machine learning.
	// +optional
	//+kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Machine Learning server URL.
	// +optional
	ServerURL *string `json:"serverURL,omitempty"`

	// Search for images semantically using CLIP embeddings.
	// If disabled, images will not be encoded for smart search.
	// +optional
	//+kubebuilder:default=true
	SmartSearchEnabled *bool `json:"smartSearchEnabled,omitempty"`

	// The name of a CLIP model listed at https://huggingface.co/immich-app .
	// Note that you must re-run the 'Smart Search' job for all images upon changing a model.
	// +kubebuilder:default="ViT-B-16-SigLIP-512__webli"
	// +optional
	SmartSearchClipModelName *string `json:"smartSearchClipModelName,omitempty"`

	// Use CLIP embeddings to find likely duplicates.
	// If disabled, exactly identical assets will still be de-duplicated.
	// +optional
	//+kubebuilder:default=true
	DuplicateDetectionEnabled *bool `json:"duplicateDetectionEnabled,omitempty"`

	// Maximum distance between two images to consider them duplicates, ranging from 0.001-0.1.
	// Higher values will detect more duplicates, but may result in false positives.
	// +kubebuilder:default="0.01"
	// +optional
	// TODO Convert from string to float
	DuplicateMaxDistance *string `json:"duplicateMaxDistance,omitempty"`

	// Detect, recognize and group faces in images.
	// If disabled, images will not be encoded for facial recognition and will not populate the People section in the Explore page.
	// +optional
	//+kubebuilder:default=true
	FacialRecognitionEnabled *bool `json:"facialRecognitionEnabled,omitempty"`

	// Facial recognition model.
	// Models are listed in descending order of size. Larger models are slower and use more memory, but produce better results.
	// Note that you must re-run the Face Detection job for all images upon changing a model.
	// +kubebuilder:validation:Enum=antelopev2;buffalo_l;buffalo_m;buffalo_s
	// +kubebuilder:default="buffalo_l"
	// +optional
	FacialRecognitionModelName *string `json:"facialRecognitionModelName,omitempty"`

	// Minimum confidence score for a face to be detected from 0-1.
	// Lower values will detect more faces but may result in false positives.
	// +kubebuilder:default="0.3"
	// +optional
	// TODO Convert from string to float
	FacialRecognitionMinScore *string `json:"facialRecognitionMinScore,omitempty"`

	// Maximum distance between two faces to be considered the same person, ranging from 0-2.
	// Lowering this can prevent labeling two people as the same person,
	// while raising it can prevent labeling the same person as two different people.
	// Note that it is easier to merge two people than to split one person in two,
	// so err on the side of a lower threshold when possible.
	// +kubebuilder:default="0.2"
	// +optional
	// TODO Convert from string to float
	FacialRecognitionMaxDistance *string `json:"facialRecognitionMaxDistance,omitempty"`

	// The minimum number of recognized faces for a person to be created.
	// Increasing this makes Facial Recognition more precise at the cost of increasing
	// the chance that a face is not assigned to a person.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	// +optional
	FacialRecognitionMinFaces *int32 `json:"facialRecognitionMinFaces,omitempty"`
}

type Logging struct {
	// Enable logging.
	// +optional
	//+kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Log level when enabled.
	// +kubebuilder:validation:Enum=fatal;error;warn;log;debug;verbose
	// +kubebuilder:default=log
	// +optional
	Level *string `json:"level,omitempty"`
}

type ExternalLibraryScanningSettings struct {

	// Enable periodic library scanning.
	// +optional
	//+kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Scanning interval using the CRON format. For more information, please refer to e.g. https://crontab.guru/
	// +optional
	//+kubebuilder:default="0 2 * * *"
	Cron *string `json:"cron,omitempty"`
}

type ExperimentalSettings struct {
	// Automatically watch external library for changed files
	// +optional
	//+kubebuilder:default=false
	WatchExternalLibrary *bool `json:"watchExternalLibrary,omitempty"`

	// +optional
	HardwareAcceleration *HardwareAcceleration `json:"hardwareAcceleration,omitempty"`
}

type HardwareAcceleration struct {

	// Acceleration API. The API that will interact with your device to accelerate transcoding.
	// This setting is 'best effort': it will fallback to software transcoding on failure.
	// VP9 may or may not work depending on your hardware.
	// nvenc means "NVENC (requires NVIDIA GPU)".
	// qsv means "Quick Sync (requires 7th gen Intel CPU or later)".
	// vaapi means "VAAPI".
	// rkmpp means "RKMPP (only on Rockchip SOCs)".
	// +kubebuilder:validation:Enum="nvenc";"qsv";"vaapi";"rkmpp";"disabled"
	// +kubebuilder:default=disabled
	// +optional
	AccelerationAPI *string `json:"accelerationAPI,omitempty"`

	// Enables end-to-end acceleration instead of only accelerating encoding.
	// May not work on all videos.
	// +optional
	//+kubebuilder:default=false
	HardwareDecodingEnabled *bool `json:"hardwareDecodingEnabled,omitempty"`

	// Constant quality mode.
	// ICQ is better than CQP, but some hardware acceleration devices do not support this mode.
	// Setting this option will prefer the specified mode when using quality-based encoding.
	// Ignored by NVENC as it does not support ICQ.
	// +kubebuilder:validation:Enum="auto";"icq";"cqp"
	// +kubebuilder:default=auto
	// +optional
	ConstantQualityMode *string `json:"constantQualityMode,omitempty"`

	// Applies only to NVENC. Increases quality of high-detail, low-motion scenes.
	// May not be compatible with older devices.
	// +optional
	//+kubebuilder:default=false
	TemporalAQEnabled *bool `json:"temporalAQEnabled,omitempty"`

	// Preferred Hardware device.
	// Applies only to VAAPI and QSV. Sets the dri node used for hardware transcoding.
	// +kubebuilder:default=auto
	// +optional
	PreferredHardwareDevice *string `json:"preferredHardwareDevice,omitempty"`
}

type ImageSettings struct {

	// Thumbnail format.
	// WebP produces smaller files than JPEG, but is slower to encode.
	// +kubebuilder:validation:Enum=webp;jpeg
	// +kubebuilder:default=webp
	// +optional
	ThumbnailFormat *string `json:"thumbnailFormat,omitempty"`

	// Thumbnail resolution.
	// Higher resolutions can preserve more detail but take longer to encode,
	// have larger file sizes and can reduce app responsiveness.
	// +kubebuilder:validation:Enum=200;250;480;720;1080
	// +kubebuilder:default=250
	// +optional
	ThumbnailResolution *int32 `json:"thumbnailResolution,omitempty"`

	// Thumbnail quality from 1-100.
	// Higher is better, but produces larger files and can reduce app responsiveness.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=80
	// +optional
	ThumbnailQuality *int32 `json:"thumbnailQuality,omitempty"`

	// Preview format.
	// WebP produces smaller files than JPEG, but is slower to encode.
	// +kubebuilder:validation:Enum=webp;jpeg
	// +kubebuilder:default=jpeg
	// +optional
	PreviewFormat *string `json:"previewFormat,omitempty"`

	// Preview resolution.
	// Higher resolutions can preserve more detail but take longer to encode,
	// have larger file sizes and can reduce app responsiveness.
	// +kubebuilder:validation:Enum=720;1080;1440;4000
	// +kubebuilder:default=1440
	// +optional
	PreviewResolution *int32 `json:"previewResolution,omitempty"`

	// Preview quality from 1-100.
	// Higher is better, but produces larger files and can reduce app responsiveness.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=80
	// +optional
	PreviewQuality *int32 `json:"previewQuality,omitempty"`

	// Prefer wide gamut.
	// Use Display P3 for thumbnails. This better preserves the vibrance of images with wide colorspaces,
	// but images may appear differently on old devices with an old browser version.
	// sRGB images are kept as sRGB to avoid color shifts.
	//+kubebuilder:default=true
	// +optional
	P3ColorSpaceEnabled *bool `json:"p3ColorSpaceEnabled,omitempty"`

	// Prefer embedded preview.
	// Use embedded previews in RAW photos as the input to image processing when available.
	// This can produce more accurate colors for some images,
	// but the quality of the preview is camera-dependent
	// and the image may have more compression artifacts.
	//+kubebuilder:default=false
	// +optional
	EmbeddedPreviewEnabled *bool `json:"embeddedPreviewEnabled,omitempty"`
}

type VideoTranscodingSettings struct {

	// Video quality level. Typical values are 23 for H.264, 28 for HEVC, 31 for VP9, and 35 for AV1.
	// Lower is better, but produces larger files.
	// +kubebuilder:default=23
	// +optional
	ConstantRateFactor *int32 `json:"constantRateFactor,omitempty"`

	// Compression speed.
	// Slower presets produce smaller files, and increase quality when targeting a certain bitrate.
	// VP9 ignores speeds above 'faster'.
	// +kubebuilder:validation:Enum=ultrafast;superfast;veryfast;faster;fast;medium;slow;slower;veryslow
	// +kubebuilder:default=ultrafast
	// +optional
	CompressionSpeed *string `json:"compressionSpeed,omitempty"`

	// Video codec.
	// VP9 has high efficiency and web compatibility, but takes longer to transcode.
	// HEVC performs similarly, but has lower web compatibility.
	// H.264 is widely compatible and quick to transcode, but produces much larger files.
	// AV1 is the most efficient codec but lacks support on older devices.
	// +kubebuilder:default={"name":"h264"}
	// +optional
	VideoCodec *VideoCodec `json:"videoCodec,omitempty"`

	// Audio codec.
	// Opus is the highest quality option, but has lower compatibility with old devices or software.
	// +kubebuilder:default={"name":"aac"}
	// +optional
	AudioCodec *AudioCodec `json:"audioCodec,omitempty"`

	// List of video codecs which do not need to be transcoded. Only used for certain transcode policies.
	// +kubebuilder:default={"name":"h264"}
	// +optional
	AcceptedVideoCodecs *[]VideoCodec `json:"acceptedVideoCodecs,omitempty"`

	// List of video codecs which do not need to be transcoded. Only used for certain transcode policies.
	// +kubebuilder:default={"name":"aac","name":"mp3","name":"libopus","name":"pcm_s16le"}
	// +optional
	AcceptedAudioCodecs *[]AudioCodec `json:"acceptedAudioCodecs,omitempty"`

	// List of container formats which do not need to be remuxed to MP4. Only used for certain transcode policies.
	// +kubebuilder:default={"name":"mov","name":"ogg","name":"webm"}
	// +optional
	AcceptedContainers *[]VideoContainer `json:"acceptedContainers,omitempty"`

	// Target resolution.
	// Higher resolutions can preserve more detail but take longer to encode,
	// have larger file sizes, and can reduce app responsiveness.
	// +kubebuilder:validation:Enum="original";"480";"720";"1080";"1440";"2160"
	// +kubebuilder:default="720"
	// +optional
	TargetResolution *string `json:"targetResolution,omitempty"`

	// Maximum bitrate.
	// Setting a max bitrate can make file sizes more predictable at a minor cost to quality.
	// At 720p, typical values are 2600k for VP9 or HEVC, or 4500k for H.264.
	// Disabled if set to 0.
	// +kubebuilder:default="0"
	// +optional
	MaxBitrate *string `json:"maxBitrate,omitempty"`

	// Higher values lead to faster encoding, but leave less room for the server to process other tasks while active.
	// This value should not be more than the number of CPU cores.
	// Maximizes utilization if set to 0.
	// +kubebuilder:default=0
	// +optional
	Threads *int32 `json:"threads,omitempty"`

	// Policy for when a video should be transcoded.
	// HDR videos will always be transcoded (except if transcoding is disabled).
	// "all" means "All videos".
	// "optimal" means "Videos higher than target resolution or not in an accepted format".
	// "bitrate" means "Videos higher than max bitrate or not in an accepted format".
	// "required" means "Only videos not in an accepted format".
	// "disabled" means "Don't transcode any videos, may break playback on some clients".
	// +kubebuilder:validation:Enum="all";"optimal";"bitrate";"required";"disabled"
	// +kubebuilder:default=required
	// +optional
	TranscodePolicy *string `json:"transcodePolicy,omitempty"`

	// Attempts to preserve the appearance of HDR videos when converted to SDR.
	// Each algorithm makes different tradeoffs for color, detail and brightness.
	// Hable preserves detail, Mobius preserves color, and Reinhard preserves brightness.
	// +kubebuilder:validation:Enum="hable";"mobius";"reinhard";"disabled"
	// +kubebuilder:default=hable
	// +optional
	ToneMappingAlgorithm *string `json:"toneMappingAlgorithm,omitempty"`

	// Transcode in two passes to produce better encoded videos.
	// When max bitrate is enabled (required for it to work with H.264 and HEVC),
	// this mode uses a bitrate range based on the max bitrate and ignores CRF.
	// For VP9, CRF can be used if max bitrate is disabled.
	// +optional
	// +kubebuilder:default=false
	TwoPassEncodingEnabled *bool `json:"twoPassEncodingEnabled,omitempty"`

	// Maximum B-frames.
	// Higher values improve compression efficiency, but slow down encoding.
	// May not be compatible with hardware acceleration on older devices.
	// 0 disables B-frames, while -1 sets this value automatically.
	// +kubebuilder:default=-1
	// +optional
	MaximumBFrames *int32 `json:"maximumBFrames,omitempty"`

	// The number of frames to reference when compressing a given frame.
	// Higher values improve compression efficiency, but slow down encoding.
	// 0 sets this value automatically.
	// +kubebuilder:default=0
	// +optional
	ReferenceFrames *int32 `json:"referenceFrames,omitempty"`

	// the maximum frame distance between keyframes.
	// Lower values worsen compression efficiency,
	// but improve seek times and may improve quality in scenes with fast movement.
	// 0 sets this value automatically.
	// +kubebuilder:default=0
	// +optional
	MaximumKFrameInterval *int32 `json:"maximumKFrameInterval,omitempty"`
}

type VideoCodec struct {
	// +kubebuilder:validation:Enum=h264;hevc;vp9;av1
	Name string `json:"name"`
}

type AudioCodec struct {
	// +kubebuilder:validation:Enum=aac;mp3;libopus;pcm_s16le
	Name string `json:"name"`
}

type VideoContainer struct {
	// +kubebuilder:validation:Enum=mov;ogg;webm
	Name string `json:"name"`
}

type OAuthLogin struct {

	// Login with OAuthLogin.
	// +optional
	//+kubebuilder:default=false
	Enabled *bool `json:"enabled,omitempty"`

	// Issuer URL
	// +optional
	IssuerUrl *string `json:"issuerUrl,omitempty"`

	// Secret containing the client ID
	// +optional
	ClientIdSecretKeyRef *SecretKeyRef `json:"clientIdSecret,omitempty"`

	// Secret containing the client ID
	// +optional
	ClientSecretSecretKeyRef *SecretKeyRef `json:"clientSecretSecret,omitempty"`

	// OAuthLogin Scopes
	// +optional
	//+kubebuilder:default={"openid","email","profile"}
	Scope *[]string `json:"scope,omitempty"`

	// Signing algorithm
	// +optional
	//+kubebuilder:default=RS256
	//TODO: list
	SigningAlgorithm *string `json:"signingAlgorithm,omitempty"`

	// Algorithm used to sign the user profile.
	// +optional
	//+kubebuilder:default=none
	//TODO: list
	ProfileSigningAlgorithm *string `json:"profileSigningAlgorithm,omitempty"`

	// Automatically set the user's storage label to the value of this claim.
	// +optional
	//+kubebuilder:default=preferred_username
	StorageLabelClaim *string `json:"storageLabelClaim,omitempty"`

	// Automatically set the user's storage quota to the value of this claim.
	// +optional
	//+kubebuilder:default=immich_quota
	StorageQuotaClaim *string `json:"storageQuotaClaim,omitempty"`

	// Quota in GiB to be used when no claim is provided (Enter 0 for unlimited quota).
	// +optional
	//+kubebuilder:default=0
	DefaultStorageQuota *int `json:"defaultStorageQuota,omitempty"`

	// Button Text
	// +optional
	//+kubebuilder:default="Login with OAuthLogin"
	ButtonText *string `json:"buttonText,omitempty"`

	// Automatically register new users after signing in with OAuthLogin
	// +optional
	//+kubebuilder:default=true
	AutoRegister *bool `json:"autoRegister,omitempty"`

	// Start the OAuthLogin login flow automatically upon navigating to the login page
	// +optional
	//+kubebuilder:default=false
	AutoLaunch *bool `json:"autoLaunch,omitempty"`

	// Enable when OAuthLogin provider does not allow a mobile URI, like {callback}
	// +optional
	MobileRedirectUri *string `json:"mobileRedirectUri,omitempty"`
}

type SecretKeyRef struct {
	// Secret Name
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// Key in the Secret
	//+kubebuilder:validation:Required
	Key string `json:"key"`
}

type DatabaseBackup struct {

	// Enable database backups.
	// +optional
	//+kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Scanning interval using the CRON format. For more information, please refer to e.g. https://crontab.guru/
	// +optional
	//+kubebuilder:default="0 02 * * *"
	Cron *string `json:"cron,omitempty"`

	// Amount of previous backups to keep
	// +optional
	//+kubebuilder:default=14
	Amount *int32 `json:"amount,omitempty"`
}

type JobConcurrency struct {

	//Thumbnail generation concurrency
	// +optional
	//+kubebuilder:default=5
	Thumbnail *int32 `json:"thumbnail,omitempty"`

	// Metadata extraction concurrency
	// +optional
	//+kubebuilder:default=5
	MetadataExtraction *int32 `json:"metadataExtraction,omitempty"`

	// Library concurrency
	// +optional
	//+kubebuilder:default=5
	Library *int32 `json:"library,omitempty"`

	// Sidecar metadata concurrency
	// +optional
	//+kubebuilder:default=5
	Sidecar *int32 `json:"sidecar,omitempty"`

	// Smart search concurrency
	// +optional
	//+kubebuilder:default=2
	SmartSearch *int32 `json:"smartSearch,omitempty"`

	// Face detection concurrency
	// +optional
	//+kubebuilder:default=2
	FaceDetection *int32 `json:"faceDetection,omitempty"`

	// Video transcoding concurrency
	// +optional
	//+kubebuilder:default=1
	VideoConversion *int32 `json:"videoConversion,omitempty"`

	// Migration concurrency
	// +optional
	//+kubebuilder:default=5
	Migration *int32 `json:"migration,omitempty"`
}

// ImmichStatus defines the observed state of Immich
type ImmichStatus struct {
	// Represents the observations of a Immich's current state.
	// Immich.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// Immich.status.conditions.status are one of True, False, Unknown.
	// Immich.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// Immich.status.conditions.Message is a human-readable message indicating details about the transition.
	// For further information see: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// Conditions store the status conditions of the Immich instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Immich is the Schema for the immiches API
type Immich struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImmichSpec   `json:"spec,omitempty"`
	Status ImmichStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImmichList contains a list of Immich
type ImmichList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Immich `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Immich{}, &ImmichList{})
}
