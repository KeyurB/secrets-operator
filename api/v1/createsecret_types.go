/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretType describes the type of secret.
// Only one of the following SecretType policies may be specified.
// +kubebuilder:validation:Enum=Opaque;Transition;Test
type SecretType string

const (
	// Opaque Secret Type.
	Opaque SecretType = "Opaque"

	// Transition secret type
	Transition SecretType = "Transition"

	// Test Secret type
	Test SecretType = "Test"
)

// CreateSecretSpec defines the desired state of CreateSecret
type CreateSecretSpec struct {
	// Specifies the type of secret.
	// Valid values are:
	// - "Opaque": Opaque Secret Type;
	// - "Transition": Transition secret type;
	// - "Test": Test Secret type
	// +kubebuilder:validation:Required
	SecretType SecretType `json:"secretType"`

	// SecretName defines the name of secret
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`

	// SecretLength define the length of secret
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=8
	SecretLength int `json:"secretLength,omitempty"`
}

// CreateSecretStatus defines the observed state of CreateSecret
type CreateSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// SecretName defines the name of secret
	SecretName string `json:"secretName,omitempty"`
	// Last SecretCreationTime defines when the secret was created
	SecretCreationTime string `json:"secretCreationTime,omitempty"`

	// SecretLength define the length of secret
	SecretLength int `json:"secretLength,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CreateSecret is the Schema for the createsecrets API
// +kubebuilder:subresource:status
type CreateSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CreateSecretSpec   `json:"spec,omitempty"`
	Status CreateSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CreateSecretList contains a list of CreateSecret
type CreateSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CreateSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CreateSecret{}, &CreateSecretList{})
}
