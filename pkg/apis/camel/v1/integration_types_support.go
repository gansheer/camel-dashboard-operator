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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// IntegrationLabel is used to tag k8s object created by a given Integration.
	IntegrationLabel = "camel.apache.org/integration"
	// IntegrationGenerationLabel is used to check on outdated integration resources that can be removed by garbage collection.
	IntegrationGenerationLabel = "camel.apache.org/generation"
	// IntegrationSyntheticLabel is used to tag k8s synthetic Integrations.
	IntegrationSyntheticLabel = "camel.apache.org/is-synthetic"
	// IntegrationImportedKindLabel specifies from what kind of resource an Integration was imported.
	IntegrationImportedKindLabel = "camel.apache.org/imported-from-kind"
	// IntegrationImportedNameLabel specifies from what resource an Integration was imported.
	IntegrationImportedNameLabel = "camel.apache.org/imported-from-name"

	// IntegrationFlowEmbeddedSourceName --.
	IntegrationFlowEmbeddedSourceName = "camel-k-embedded-flow.yaml"
)

func NewIntegration(namespace string, name string) Integration {
	return Integration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       IntegrationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func NewIntegrationList() IntegrationList {
	return IntegrationList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       IntegrationKind,
		},
	}
}
