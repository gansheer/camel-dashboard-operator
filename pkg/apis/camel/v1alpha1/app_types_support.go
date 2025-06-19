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
	// AppLabel is used to tag k8s object created by a given Camel Application.
	AppLabel = "camel.apache.org/app"
	// AppSyntheticLabel is used to tag k8s synthetic Camel Applications.
	AppSyntheticLabel = "camel.apache.org/is-synthetic"
	// AppImportedKindLabel specifies from what kind of resource an App was imported.
	AppImportedKindLabel = "camel.apache.org/imported-from-kind"
	// AppImportedNameLabel specifies from what resource an App was imported.
	AppImportedNameLabel = "camel.apache.org/imported-from-name"
	// AppPollingIntervalSecondsAnnotation is used to instruct a given application to poll interval.
	AppPollingIntervalSecondsAnnotation = "camel.apache.org/polling-interval-seconds"
	// AppObservabilityServicesPort is used to instruct an application to use a specific port for metrics scraping.
	AppObservabilityServicesPort = "camel.apache.org/observability-services-port"
)

func NewApp(namespace string, name string) App {
	return App{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       AppKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func NewAppList() AppList {
	return AppList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       AppKind,
		},
	}
}

func (appStatus *AppStatus) AddCondition(condition metav1.Condition) {
	if appStatus.Conditions == nil {
		appStatus.Conditions = []metav1.Condition{}
	}
	appStatus.Conditions = append(appStatus.Conditions, condition)
}
