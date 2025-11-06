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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// Image is the container image to run
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Command to run in the container (overrides image entrypoint)
	// +optional
	Command []string `json:"command,omitempty"`

	// Args to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`

	// Environment variables for the function
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// Number of replicas for the function deployment
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resource requirements for the function
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// ResourceRequirements defines resource requests and limits
type ResourceRequirements struct {
	// CPU request (e.g., "100m")
	// +optional
	CPURequest string `json:"cpuRequest,omitempty"`

	// Memory request (e.g., "128Mi")
	// +optional
	MemoryRequest string `json:"memoryRequest,omitempty"`

	// CPU limit (e.g., "200m")
	// +optional
	CPULimit string `json:"cpuLimit,omitempty"`

	// Memory limit (e.g., "256Mi")
	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`
}

// FunctionStatus defines the observed state of Function.
type FunctionStatus struct {
	// Phase represents the current lifecycle phase of the function
	// +kubebuilder:validation:Enum=Pending;Running;Failed;Unknown
	// +optional
	Phase string `json:"phase,omitempty"`

	// Replicas is the number of desired replicas
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// AvailableReplicas is the number of available replicas
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// LastUpdated is the timestamp of the last status update
	// +optional
	LastUpdated string `json:"lastUpdated,omitempty"`

	// Conditions represent the current state of the Function resource
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Function is the Schema for the functions API
type Function struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Function
	// +required
	Spec FunctionSpec `json:"spec"`

	// status defines the observed state of Function
	// +optional
	Status FunctionStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
