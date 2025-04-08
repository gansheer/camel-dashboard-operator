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
	"context"
	"fmt"
	"io"
	"net/http"

	v1alpha1 "github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// nonManagedCamelDeployment represents a regular Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelDeployment struct {
	deploy *appsv1.Deployment
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelDeployment) CamelApp(ctx context.Context, c client.Client) *v1alpha1.App {
	newApp := v1alpha1.NewApp(app.deploy.Namespace, app.deploy.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.deploy.Name,
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       app.deploy.Name,
			UID:        app.deploy.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)
	// TODO: we should expect this to be correctly handled by a proper reconciliation cycle instead
	deployImage := app.GetAppImage()
	appPhase := app.GetAppPhase()
	newApp.Status.Phase = appPhase
	newApp.Status.Image = deployImage
	pods, err := app.GetPods(ctx, c)
	if err != nil {
		return nil
	}
	newApp.Status.Pods = pods
	newApp.Status.Replicas = app.deploy.Spec.Replicas
	return &newApp
}

// GetAppPhase returns the phase of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppPhase() v1alpha1.AppPhase {
	if app.deploy.Status.AvailableReplicas == app.deploy.Status.Replicas {
		return v1alpha1.AppPhaseRunning
	}

	return v1alpha1.AppPhaseError
}

// GetAppImage returns the container image of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppImage() string {
	return app.deploy.Spec.Template.Spec.Containers[0].Image
}

// GetPods returns the pods backing the Camel application.
func (app *nonManagedCamelDeployment) GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error) {
	var podsInfo []v1alpha1.PodInfo
	pods := &corev1.PodList{}
	err := c.List(ctx, pods,
		ctrl.InNamespace(app.deploy.GetNamespace()),
		ctrl.MatchingLabels(app.deploy.Spec.Selector.MatchLabels),
	)
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		podIp := pod.Status.PodIP
		podInfo := v1alpha1.PodInfo{
			Name:       pod.GetName(),
			Status:     string(pod.Status.Phase),
			InternalIP: podIp,
		}
		setMetrics(&podInfo, podIp)
		setHealth(&podInfo, podIp)

		podsInfo = append(podsInfo, podInfo)
	}

	return podsInfo, nil
}

func setMetrics(podInfo *v1alpha1.PodInfo, podIp string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", podIp, observabilityPortDefault, observabilityMetricsDefault))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		if podInfo.ObservabilityService == nil {
			podInfo.ObservabilityService = &v1alpha1.ObservabilityServiceInfo{}
		}
		podInfo.ObservabilityService.MetricsEndpoint = observabilityMetricsDefault
		podInfo.ObservabilityService.MetricsPort = observabilityPortDefault

		metrics, err := parseMetrics(resp.Body)
		if err != nil {
			return err
		}
		if metric, ok := metrics["app_info"]; ok {
			populateRuntimeInfo(metric, podInfo)
		}
		if metric, ok := metrics["camel_exchanges_total"]; ok {
			populateRuntimeExchangeTotal(metric, podInfo)
		}
		if metric, ok := metrics["camel_exchanges_failed_total"]; ok {
			populateRuntimeExchangeFailedTotal(metric, podInfo)
		}
		if metric, ok := metrics["camel_exchanges_succeeded_total"]; ok {
			populateRuntimeExchangeSucceededTotal(metric, podInfo)
		}
		if metric, ok := metrics["camel_exchanges_inflight"]; ok {
			populateRuntimeExchangeInflight(metric, podInfo)
		}
	}

	return nil
}

func parseMetrics(reader io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return mf, nil
}

func populateRuntimeInfo(metric *dto.MetricFamily, podInfo *v1alpha1.PodInfo) {
	if podInfo.Runtime == nil {
		podInfo.Runtime = &v1alpha1.RuntimeInfo{}
	}
	if podInfo.Runtime.Exchange == nil {
		podInfo.Runtime.Exchange = &v1alpha1.ExchangeInfo{}
	}
	for _, label := range metric.GetMetric()[0].GetLabel() {
		switch ptr.Deref(label.Name, "") {
		case "camel_runtime_provider":
			podInfo.Runtime.RuntimeProvider = ptr.Deref(label.Value, "")
		case "camel_runtime_version":
			podInfo.Runtime.RuntimeVersion = ptr.Deref(label.Value, "")
		case "camel_version":
			podInfo.Runtime.CamelVersion = ptr.Deref(label.Value, "")
		}
	}
}

func populateRuntimeExchangeTotal(metric *dto.MetricFamily, podInfo *v1alpha1.PodInfo) {
	podInfo.Runtime.Exchange.Total = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateRuntimeExchangeFailedTotal(metric *dto.MetricFamily, podInfo *v1alpha1.PodInfo) {
	podInfo.Runtime.Exchange.Failed = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateRuntimeExchangeSucceededTotal(metric *dto.MetricFamily, podInfo *v1alpha1.PodInfo) {
	podInfo.Runtime.Exchange.Succeeded = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateRuntimeExchangeInflight(metric *dto.MetricFamily, podInfo *v1alpha1.PodInfo) {
	podInfo.Runtime.Exchange.Pending = int(ptr.Deref(metric.GetMetric()[0].GetGauge().Value, 0))
}

func setHealth(podInfo *v1alpha1.PodInfo, podIp string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", podIp, observabilityPortDefault, observabilityHealthDefault))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		if podInfo.ObservabilityService == nil {
			podInfo.ObservabilityService = &v1alpha1.ObservabilityServiceInfo{}
		}
		podInfo.ObservabilityService.HealthPort = observabilityPortDefault
		podInfo.ObservabilityService.HealthEndpoint = observabilityHealthDefault

		status, err := parseHealthStatus(resp.Body)
		if err != nil {
			return err
		}
		if podInfo.Runtime == nil {
			podInfo.Runtime = &v1alpha1.RuntimeInfo{}
		}
		podInfo.Runtime.Status = status
	}

	return nil
}

func parseHealthStatus(reader io.Reader) (string, error) {
	// TODO: parsing logic here!
	return "up", nil
}
