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
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/squakez/camel-dashboard-operator/pkg/util/log"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	OperatorWatchNamespaceEnvVariable = "WATCH_NAMESPACE"
	operatorNamespaceEnvVariable      = "NAMESPACE"
	operatorPodNameEnvVariable        = "POD_NAME"
)

const OperatorLockName = "camel-k-lock"

var OperatorImage string

// IsCurrentOperatorGlobal returns true if the operator is configured to watch all namespaces.
func IsCurrentOperatorGlobal() bool {
	if watchNamespace, envSet := os.LookupEnv(OperatorWatchNamespaceEnvVariable); !envSet || strings.TrimSpace(watchNamespace) == "" {
		log.Debug("Operator is global to all namespaces")
		return true
	}

	log.Debug("Operator is local to namespace")
	return false
}

// GetOperatorPod returns the Pod which is running the operator in a given namespace.
func GetOperatorPod(ctx context.Context, c ctrl.Reader, ns string) *corev1.Pod {
	lst := corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
	}
	if err := c.List(ctx, &lst,
		ctrl.InNamespace(ns),
		ctrl.MatchingLabels{
			"camel.apache.org/component": "operator",
		}); err != nil {
		return nil
	}
	if len(lst.Items) == 0 {
		return nil
	}
	return &lst.Items[0]
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

// GetOperatorPodName returns the pod that is running the current operator (if any).
func GetOperatorPodName() string {
	if podName, envSet := os.LookupEnv(operatorPodNameEnvVariable); envSet {
		return podName
	}
	return ""
}

// GetOperatorLockName returns the name of the lock lease that is electing a leader on the particular namespace.
func GetOperatorLockName(operatorID string) string {
	return fmt.Sprintf("%s-lock", operatorID)
}

// FilteringFuncs do preliminary checks to determine if certain events should be handled by the controller
// based on labels on the resources (e.g. camel.apache.org/operator.id) and the operator configuration,
// before handing the computation over to the user code.
type FilteringFuncs[T ctrl.Object] struct {
	// Create returns true if the Create event should be processed
	CreateFunc func(event.TypedCreateEvent[T]) bool

	// Delete returns true if the Delete event should be processed
	DeleteFunc func(event.TypedDeleteEvent[T]) bool

	// Update returns true if the Update event should be processed
	UpdateFunc func(event.TypedUpdateEvent[T]) bool

	// Generic returns true if the Generic event should be processed
	GenericFunc func(event.TypedGenericEvent[T]) bool
}
