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

// Package v1 contains a minimal subset of the OpenShift Console API types
// needed by the operator to create ConsolePlugin resources. Defined locally
// to avoid pulling in github.com/openshift/api which conflicts with the
// project's k8s dependency versions.
package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	GroupName = "console.openshift.io"
	Version   = "v1"
)

type ConsolePluginBackendType string

const ServiceBackendType ConsolePluginBackendType = "Service"

type LoadType string

const PreloadType LoadType = "Preload"

type ConsolePlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ConsolePluginSpec `json:"spec"`
}

type ConsolePluginSpec struct {
	DisplayName string               `json:"displayName"`
	Backend     ConsolePluginBackend `json:"backend"`
	I18n        ConsolePluginI18n    `json:"i18n"`
}

type ConsolePluginBackend struct {
	Type    ConsolePluginBackendType `json:"type"`
	Service *ConsolePluginService    `json:"service,omitempty"`
}

type ConsolePluginService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Port      int32  `json:"port"`
	BasePath  string `json:"basePath,omitempty"`
}

type ConsolePluginI18n struct {
	LoadType LoadType `json:"loadType"`
}

type ConsolePluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ConsolePlugin `json:"items"`
}
