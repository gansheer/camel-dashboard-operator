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

package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
)

const (
	OperatorWatchNamespaceEnvVariable = "WATCH_NAMESPACE"
	operatorNamespaceEnvVariable      = "NAMESPACE"
	CamelAppLabelSelector             = "LABEL_SELECTOR"

	CamelAppPollIntervalSeconds   = "POLL_INTERVAL_SECONDS"
	DefaultPollingIntervalSeconds = 60

	CamelAppObservabilityPort       = "OBSERVABILITY_PORT"
	defaultObservabilityPort    int = 9876
	DefaultObservabilityMetrics     = "observe/metrics"
	DefaultObservabilityHealth      = "observe/health"

	OperatorLockName = "camel-dashboard-lock"
)

// IsCurrentOperatorGlobal returns true if the operator is configured to watch all namespaces.
func IsCurrentOperatorGlobal() bool {
	if watchNamespace, envSet := os.LookupEnv(OperatorWatchNamespaceEnvVariable); !envSet || strings.TrimSpace(watchNamespace) == "" {
		log.Debug("Operator is global to all namespaces")
		return true
	}

	log.Debug("Operator is local to namespace")
	return false
}

// GetOperatorWatchNamespace returns the namespace the operator watches.
func GetOperatorWatchNamespace() string {
	if namespace, envSet := os.LookupEnv(OperatorWatchNamespaceEnvVariable); envSet {
		return namespace
	}
	return ""
}

// GetOperatorNamespace returns the namespace where the current operator is located (if set).
func GetOperatorNamespace() string {
	if podNamespace, envSet := os.LookupEnv(operatorNamespaceEnvVariable); envSet {
		return podNamespace
	}
	return ""
}

// GetOperatorLockName returns the name of the lock lease that is electing a leader on the particular namespace.
func GetOperatorLockName(operatorID string) string {
	return fmt.Sprintf("%s-lock", operatorID)
}

// GetAppLabelSelector returns the label selector used to determine a Camel application.
func GetAppLabelSelector() string {
	if labelSelector, envSet := os.LookupEnv(CamelAppLabelSelector); envSet && labelSelector != "" {
		return labelSelector
	}
	return v1alpha1.AppLabel
}

// getPollingIntervalSeconds returns the polling interval (in seconds) for the operator. It fallbacks to default value.
func getPollingIntervalSeconds() int {
	if pollingIntervalSeconds, envSet := os.LookupEnv(CamelAppPollIntervalSeconds); envSet && pollingIntervalSeconds != "" {
		interval, err := strconv.Atoi(pollingIntervalSeconds)
		if err == nil {
			return interval
		} else {
			log.Errorf(err, "could not properly parse Operator polling interval configuration, "+
				"fallback to default value %d", DefaultPollingIntervalSeconds)
		}
	}
	return DefaultPollingIntervalSeconds
}

// GetPollingInterval returns the polling interval for the operator. It fallbacks to default value.
func GetPollingInterval() time.Duration {
	return time.Duration(getPollingIntervalSeconds()) * time.Second
}

// GetObservabilityPort returns the observability por set for the operator. It fallbacks to default value.
func GetObservabilityPort() int {
	if observabilityPort, envSet := os.LookupEnv(CamelAppObservabilityPort); envSet && observabilityPort != "" {
		observabilityPortInt, err := strconv.Atoi(observabilityPort)
		if err == nil {
			return observabilityPortInt
		} else {
			log.Error(err, "could not properly parse Operator observability port configuration, "+
				"fallback to default value %d", defaultObservabilityPort)
		}
	}
	return defaultObservabilityPort
}
