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

package v2025

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PetSpec defines the desired state of Pet.
type PetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the pet
	// +kubebuilder:validation:Required
	Nickname string `json:"nickname"`

	// FoodDecayRate is the amount reduced from [PetStatus.Food]
	// +kubebuilder:default=1
	FoodDecayRate int `json:"foodDecayRate,omitempty"`

	// LoveDecayRate is the amount reduced from [PetStatus.Love]
	// +kubebuilder:default=1
	LoveDecayRate int `json:"loveDecayRate,omitempty"`

	// DecayInterval is the interval in which the love and food is decayed for this pet
	// +kubebuilder:default="10s"
	DecayInterval metav1.Duration `json:"decayInterval,omitempty"`
}

// PetStatus defines the observed state of Pet.
type PetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Food is the amount of food the pet has
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Food int `json:"food,omitempty"`

	// Love is the amount of love the pet has
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Love int `json:"love,omitempty" `

	// FedTime is the last time the pet was fed
	FedTime metav1.Time `json:"fedTime,omitempty"`

	// PetTime is the last time the pet was petted
	PetTime metav1.Time `json:"petTime,omitempty"`

	// ModifiedTime is the last time the controller modified food or love
	ModifiedTime metav1.Time `json:"modifiedTime,omitempty"`

	// Initialized
	Initialized bool `json:"initialized"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="FOOD",type=integer,JSONPath=`.status.food`
// +kubebuilder:printcolumn:name="LOVE",type=integer,JSONPath=`.status.love`
// +kubebuilder:printcolumn:name="FED_TIME",type=date,JSONPath=`.status.fedTime`
// +kubebuilder:printcolumn:name="PET_TIME",type=date,JSONPath=`.status.petTime`
// +kubebuilder:printcolumn:name="MODIFIED_TIME",type=date,JSONPath=`.status.modifiedTime`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Pet is the Schema for the pets API.
type Pet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PetSpec   `json:"spec,omitempty"`
	Status PetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PetList contains a list of Pet.
type PetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pet{}, &PetList{})
}
