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

	v1alpha1 "github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
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
	// TODO: all the code above is (still) a simulation.
	// We should expect this to be correctly handled by a proper reconciliation cycle instead
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
		podInfo := v1alpha1.PodInfo{
			Name:       pod.GetName(),
			Status:     string(pod.Status.Phase),
			InternalIP: pod.Status.PodIP,
		}

		// TODO: scan metrics here
		podsInfo = append(podsInfo, podInfo)
	}

	return podsInfo, nil
}
