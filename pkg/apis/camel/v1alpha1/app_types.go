/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

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

const (
	// IntegrationKind --.
	IntegrationKind string = "Integration"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=apps,scope=Namespaced,shortName=capp,categories=camel
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// App is the Schema for the Camel Applications API.
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// the desired Integration specification
	Spec AppSpec `json:"spec,omitempty"`
	// the status of the Integration
	Status AppStatus `json:"status,omitempty"`
}

// AppSpec specifies the configuration of an App.
type AppSpec struct {
}

// AppStatus defines the observed state of an App.
type AppStatus struct {
	// the actual phase
	Phase AppPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true

// AppList contains a list of Apps.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

// AppPhase --.
type AppPhase string
