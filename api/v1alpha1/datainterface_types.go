/*
Copyright 2023.

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

// DataInterfaceSpec defines the desired state of DataInterface
type DataInterfaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the data interface
	Name string `json:"name,omitempty"`
	// Reference is a clusterwide unique identifier for the data interface
	// if empty, the name will be used as reference instead
	// +optional
	Reference *string `json:"reference,omitempty"`
	// Type is the type of the data interface (e.g. "topic", "queue", "database", "file")
	Type string `json:"type,omitempty"`
	// Description is a human-readable description of the data interface
	// +optional
	Description *string `json:"description,omitempty"`
}

// DataInterfaceStatus defines the observed state of DataInterface
type DataInterfaceStatus struct {
	// UsedReferences is the generated name of the data interface
	UsedReference string `json:"usedReference,omitempty"`
	// UsedInDataProcesses is a list of data processes that use this data interface
	UsedInDataProcesses []metav1.OwnerReference `json:"usedInDataProcesses,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataInterface is the Schema for the datainterfaces API
type DataInterface struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataInterfaceSpec   `json:"spec,omitempty"`
	Status DataInterfaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataInterfaceList contains a list of DataInterface
type DataInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataInterface `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataInterface{}, &DataInterfaceList{})
}
