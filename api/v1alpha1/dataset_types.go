/*
Copyright 2023 Florian Schrag.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataSetSpec defines the desired state of DataSet
type DataSetSpec struct {
	Interfaces []DataInterfaceSpec `json:"interfaces,omitempty"`
	Processes  []DataProcessSpec   `json:"processes,omitempty"`
}

// DataSetStatus defines the observed state of DataSet
type DataSetStatus struct {

	// A list of pointers to generated interfaces
	// +optional
	Interfaces map[string]corev1.ObjectReference `json:"interfaces,omitempty"`

	// A list of pointers to generated processes
	// +optional
	Processes map[string]corev1.ObjectReference `json:"processes,omitempty"`

	// Conditions store the status conditions of the data set
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataSet is the Schema for the datasets API it contains a set of data interfaces and data processes
type DataSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSetSpec   `json:"spec,omitempty"`
	Status DataSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataSetList contains a list of DataSet
type DataSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataSet{}, &DataSetList{})
}
