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

// Edge is a data interface that is used as input or output for a data process
type Edge struct {
	// References the data interface
	Reference string `json:"reference,omitempty"`
	// Info is a human-readable description of the data interface
	Info string `json:"info,omitempty"`
	// Trigger is true if the data interface triggers further processing
	// Outgoing edges to kafka topics usually have this set to true while
	// incoming edges from kafka topics usually have this set to true.
	Trigger bool `json:"trigger,omitempty"`
	// Description is a human-readable description of the data interface
	// +optional
	Description *string `json:"description,omitempty"`
}

// DataProcessSpec defines the desired state of DataProcess
type DataProcessSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the displayed name of the data process
	Name string `json:"name,omitempty"`

	// Type is the type of the data process (e.g. "spring-cloud-stream", "kafka-streams", "spark-streaming")
	Type string `json:"type,omitempty"`
	// Description is a human-readable description of the data process
	Description string `json:"description,omitempty"`
	// Inputs is a list of data interfaces that are used as input for the data process
	Inputs []Edge `json:"inputs,omitempty"`
	// Outputs is a list of data interfaces that are used as output for the data process
	Outputs []Edge `json:"outputs,omitempty"`
}

// DataProcessStatus defines the observed state of DataProcess
type DataProcessStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	MissingDataInterfaces []string `json:"missingDataInterfaces,omitempty"`

	// Loaded is true if the data process is loaded into the system
	Loaded bool `json:"loaded,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataProcess is the Schema for the dataprocesses API
type DataProcess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataProcessSpec   `json:"spec,omitempty"`
	Status DataProcessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataProcessList contains a list of DataProcess
type DataProcessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataProcess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataProcess{}, &DataProcessList{})
}
