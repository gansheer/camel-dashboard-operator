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

package synthetic

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
)

/*

This is a simulation used just for the POC. We'll need to wire all the pieces required to connect to the cluster and get all the information
in a real implementation.

*/

func getAppPhase() v1alpha1.AppPhase {
	if randRange(1, 5) == 1 {
		return v1alpha1.AppPhase("Error")
	}
	return v1alpha1.AppPhase("Running")
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func getPods(image string, status v1alpha1.AppPhase, podsLen *int32) []v1alpha1.PodInfo {
	var pods []v1alpha1.PodInfo
	for range *podsLen {
		pods = append(pods, getPod(image, status))
	}
	return pods
}

func getPod(image string, status v1alpha1.AppPhase) v1alpha1.PodInfo {
	pod := v1alpha1.PodInfo{
		Name:                 fmt.Sprintf("pod-%d", randRange(1000, 9999)),
		InternalIP:           fmt.Sprintf("%d.%d.%d.%d", randRange(1, 255), randRange(1, 255), randRange(1, 255), randRange(1, 255)),
		Status:               string(status),
		ObservabilityService: getObservabilityService(image),
	}

	if !strings.Contains(image, "missing") {
		pod.Runtime = getRuntimeInfo(image)
	}

	return pod
}

func getRuntimeInfo(image string) *v1alpha1.RuntimeInfo {
	totalExchange := randRange(10, 99)
	failedExchange := randRange(1, 5)
	pendingExchange := randRange(1, 3)
	succeededExchange := totalExchange - failedExchange - pendingExchange

	runtimeInfo := v1alpha1.RuntimeInfo{
		ContextName:  "camel-1",
		CamelVersion: "4.8.5",
		Exchange: v1alpha1.ExchangeInfo{
			Total:   totalExchange,
			Failed:  failedExchange,
			Pending: pendingExchange,
			Succeed: succeededExchange,
		},
	}

	if strings.Contains(image, "quarkus") {
		runtimeInfo.RuntimeProvider = "quarkus"
		runtimeInfo.RuntimeVersion = "3.18.3"
	} else if strings.Contains(image, "spring") {
		runtimeInfo.RuntimeProvider = "spring-boot"
		runtimeInfo.RuntimeVersion = "3.4.3"
	} else {
		runtimeInfo.RuntimeProvider = "main"
		runtimeInfo.RuntimeVersion = "4.8.5"
	}

	return &runtimeInfo
}

func getObservabilityService(image string) v1alpha1.ObservabilityServiceInfo {
	if strings.Contains(image, "missing") {
		return v1alpha1.ObservabilityServiceInfo{
			HealthEndpoint: "q/health",
		}
	}
	return v1alpha1.ObservabilityServiceInfo{
		HealthEndpoint:  "observe/health",
		MetricsEndpoint: "observe/metrics",
	}
}
