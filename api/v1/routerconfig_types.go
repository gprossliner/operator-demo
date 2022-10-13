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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RouterConfigSpec defines the desired state of RouterConfig
type RouterConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RouterConfig. Edit routerconfig_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	// RouterConfigName is the name of the RouterConfig. There can be 0:n Routes associated with this config
	RouterConfigName string `json:"routerConfigName,omitempty"`
}

// RouterConfigStatus defines the observed state of RouterConfig
type RouterConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// Routes is a list of all known routes.
	// This is updates by the Route-Controller
	Routes []RouteReference `json:"routes,omitempty"`
}

// RouteReference represents a single route for the RouterConfig
type RouteReference struct {
	Namespace          string `json:"namespace"`
	Name               string `json:"name"`
	ObservedGeneration int64  `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RouterConfig is the Schema for the routerconfigs API
type RouterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterConfigSpec   `json:"spec,omitempty"`
	Status RouterConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RouterConfigList contains a list of RouterConfig
type RouterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouterConfig{}, &RouterConfigList{})
}
