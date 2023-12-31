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
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Filter struct {
	// Namespace to filter for, if empty all namespaces are used
	Namespaces []string `json:"namespaces,omitempty"`
	// Labels to filter for, if empty all labels are used
	Labels map[string]string `json:"labels,omitempty"`
}

// ConstellationSpec defines the desired state of Constellation
type ConstellationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Filters is a list of filters that are used to select the data interfaces and data processes
	// If empty, all data interfaces and data processes are used
	Filters []Filter `json:"filters,omitempty"`

	// TargetConfigMap is the name of the config map that is used to store the constellation
	// it uses the same namespace as the constellation
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +required
	TargetConfigMap string `json:"targetConfigMap,omitempty"`

	// Name is the displayed name of the constellation
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +required
	Name string `json:"name,omitempty"`

	// Description is a human-readable description of the constellation
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Description *string `json:"description,omitempty"`
}

type ConstellationInterface struct {
	// Name is the displayed name of the data interface
	Name string `json:"name,omitempty"`
	// Reference is a cluster-wide unique identifier for the data interface
	Reference string `json:"reference,omitempty"`
	// Type is the type of the data interface (e.g. "topic", "queue", "database", "file")
	Type string `json:"type,omitempty"`
	// Description is a human-readable description of the data interface
	Description string `json:"description,omitempty"`
	// Labels is a set of labels for the data interface
	Labels map[string]string `json:"labels,omitempty"`
	// Source is the namespaced name of the data interface
	Source NamespacedName `json:"source,omitempty"`
}

type ConstellationEdge struct {
	// References the data interface
	Reference string `json:"reference,omitempty"`
	// Info is a human-readable description of the data interface
	Info string `json:"info,omitempty"`
	// Trigger is true if the data interface triggers further processing
	// Outgoing edges to kafka topics usually have this set to true while
	// incoming edges from kafka topics usually have this set to true.
	Trigger bool `json:"trigger,omitempty"`
	// Description is a human-readable description of the data interface
	Description string `json:"description,omitempty"`
}

type ConstellationDataProcess struct {
	// Name is the displayed name of the data process
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Name string `json:"name,omitempty"`
	// Type is the type of the data process (e.g. "spring-cloud-stream", "kafka-streams", "spark-streaming")
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Type string `json:"type,omitempty"`
	// Description is a human-readable description of the data process
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Description string `json:"description,omitempty"`
	// Inputs is a list of data interfaces that are used as input for the data process
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Inputs []ConstellationEdge `json:"inputs,omitempty"`
	// Outputs is a list of data interfaces that are used as output for the data process
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Outputs []ConstellationEdge `json:"outputs,omitempty"`
	// Labels is a set of labels for the data interface
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Labels map[string]string `json:"labels,omitempty"`
	// Source is the namespaced name of the data process
	//+operator-sdk:csv:customresourcedefinitions:type=status
	Source NamespacedName `json:"source,omitempty"`
}

type ConstellationResult struct {
	Name              string                     `json:"name,omitempty"`
	Description       string                     `json:"description,omitempty"`
	LastUpdated       metav1.Time                `json:"lastUpdated,omitempty"`
	DataInterfaceList []ConstellationInterface   `json:"dataInterfaceList"`
	DataProcessList   []ConstellationDataProcess `json:"dataProcessList"`
}

func (in *ConstellationResult) AddDataInterfaceList(log logr.Logger, items []DataInterface) {
	existing := make(map[string]bool)
	for _, item := range in.DataInterfaceList {
		existing[item.Reference] = true
	}
	for _, item := range items {
		if !existing[item.Status.UsedReference] {
			newItem := ConstellationInterface{
				Name:        item.Spec.Name,
				Reference:   item.Status.UsedReference,
				Type:        item.Spec.Type,
				Description: asString(item.Spec.Description),
				Labels:      item.Labels,
				Source: NamespacedName{
					Namespace: item.Namespace,
					Name:      item.Name,
				},
			}
			if newItem.Reference == "" {
				log.Info("Reference is empty - skipping", "item", newItem.Source)
				continue
			}
			in.DataInterfaceList = append(in.DataInterfaceList, newItem)
			existing[newItem.Reference] = true
		}
	}
}

func (in *ConstellationResult) GenerateMissingInterfaces() {
	existing := make(map[string]bool)
	for _, item := range in.DataInterfaceList {
		existing[item.Reference] = true
	}
	for _, item := range in.DataProcessList {
		for _, input := range item.Inputs {
			if !existing[input.Reference] {
				newItem := ConstellationInterface{
					Name:        input.Reference,
					Reference:   input.Reference,
					Type:        "missing",
					Description: "missing",
					Labels:      nil,
					Source: NamespacedName{
						Namespace: item.Source.Namespace,
						Name:      item.Source.Name,
					},
				}
				in.DataInterfaceList = append(in.DataInterfaceList, newItem)
				existing[newItem.Reference] = true
			}
		}
		for _, output := range item.Outputs {
			if !existing[output.Reference] {
				newItem := ConstellationInterface{
					Name:        output.Reference,
					Reference:   output.Reference,
					Type:        "missing",
					Description: "missing",
					Labels:      nil,
					Source: NamespacedName{
						Namespace: item.Source.Namespace,
						Name:      item.Source.Name,
					},
				}
				in.DataInterfaceList = append(in.DataInterfaceList, newItem)
				existing[newItem.Reference] = true
			}
		}
	}
}

func (in *ConstellationResult) AddDataProcessList(_ logr.Logger, items []DataProcess) {
	for _, item := range items {
		var inputs []ConstellationEdge
		for _, input := range item.Spec.Inputs {
			inputs = append(inputs, ConstellationEdge{
				Reference:   input.Reference,
				Info:        input.Info,
				Trigger:     input.Trigger,
				Description: asString(input.Description),
			})
		}
		var outputs []ConstellationEdge
		for _, output := range item.Spec.Outputs {
			outputs = append(outputs, ConstellationEdge{
				Reference:   output.Reference,
				Info:        output.Info,
				Trigger:     output.Trigger,
				Description: asString(output.Description),
			})
		}

		in.DataProcessList = append(in.DataProcessList, ConstellationDataProcess{
			Name:        item.Spec.Name,
			Type:        item.Spec.Type,
			Description: item.Spec.Description,
			Inputs:      inputs,
			Outputs:     outputs,
			Labels:      item.Labels,
			Source: NamespacedName{
				Namespace: item.Namespace,
				Name:      item.Name,
			},
		})
	}
}

func asString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// ConstellationStatus defines the observed state of Constellation
type ConstellationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ConstellationResult is the result of the constellation
	// +operator-sdk:csv:customresourcedefinitions:type=status
	ConstellationResult *ConstellationResult `json:"constellationResult,omitempty"`

	// Conditions store the status conditions of the constellation
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Constellation is the Schema for the constellations API
type Constellation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConstellationSpec   `json:"spec,omitempty"`
	Status ConstellationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConstellationList contains a list of Constellation
type ConstellationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Constellation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Constellation{}, &ConstellationList{})
}
