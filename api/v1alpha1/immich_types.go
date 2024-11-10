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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImmichSpec defines the desired state of Immich
type ImmichSpec struct {

	// Application settings
	Application *Application `json:"application,omitempty"`
}

type Application struct {

	// Database Backup settings.
	// +optional
	DatabaseBackup *DatabaseBackup `json:"databaseBackup,omitempty"`

	// Job concurrency settings
	// +optional
	Concurrency *Concurrency `json:"concurrency,omitempty"`
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

type Concurrency struct {

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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
