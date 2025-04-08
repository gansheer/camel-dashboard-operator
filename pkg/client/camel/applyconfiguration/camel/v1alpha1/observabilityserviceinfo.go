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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// ObservabilityServiceInfoApplyConfiguration represents a declarative configuration of the ObservabilityServiceInfo type for use
// with apply.
type ObservabilityServiceInfoApplyConfiguration struct {
	HealthEndpoint  *string `json:"healthEndpoint,omitempty"`
	HealthPort      *int    `json:"healthPort,omitempty"`
	MetricsEndpoint *string `json:"metricsEndpoint,omitempty"`
	MetricsPort     *int    `json:"metricsPort,omitempty"`
}

// ObservabilityServiceInfoApplyConfiguration constructs a declarative configuration of the ObservabilityServiceInfo type for use with
// apply.
func ObservabilityServiceInfo() *ObservabilityServiceInfoApplyConfiguration {
	return &ObservabilityServiceInfoApplyConfiguration{}
}

// WithHealthEndpoint sets the HealthEndpoint field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HealthEndpoint field is set to the value of the last call.
func (b *ObservabilityServiceInfoApplyConfiguration) WithHealthEndpoint(value string) *ObservabilityServiceInfoApplyConfiguration {
	b.HealthEndpoint = &value
	return b
}

// WithHealthPort sets the HealthPort field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the HealthPort field is set to the value of the last call.
func (b *ObservabilityServiceInfoApplyConfiguration) WithHealthPort(value int) *ObservabilityServiceInfoApplyConfiguration {
	b.HealthPort = &value
	return b
}

// WithMetricsEndpoint sets the MetricsEndpoint field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MetricsEndpoint field is set to the value of the last call.
func (b *ObservabilityServiceInfoApplyConfiguration) WithMetricsEndpoint(value string) *ObservabilityServiceInfoApplyConfiguration {
	b.MetricsEndpoint = &value
	return b
}

// WithMetricsPort sets the MetricsPort field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MetricsPort field is set to the value of the last call.
func (b *ObservabilityServiceInfoApplyConfiguration) WithMetricsPort(value int) *ObservabilityServiceInfoApplyConfiguration {
	b.MetricsPort = &value
	return b
}
