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

package app

import (
	"context"
	"fmt"

	"github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	"github.com/squakez/camel-dashboard-operator/pkg/controller/synthetic"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMonitorAction returns an action that monitors the App.
func NewMonitorAction() Action {
	return &monitorAction{}
}

type monitorAction struct {
	baseAction
}

func (action *monitorAction) Name() string {
	return "monitor"
}

func (action *monitorAction) CanHandle(app *v1alpha1.App) bool {
	// It always apply, regardless the phase
	return true
}

func (action *monitorAction) Handle(ctx context.Context, app *v1alpha1.App) (*v1alpha1.App, error) {
	action.L.Infof("Monitoring App %s/%s", app.Namespace, app.Name)
	objOwner, err := lookupObject(ctx, action.client,
		app.Annotations[v1alpha1.AppImportedKindLabel], app.Namespace, app.Annotations[v1alpha1.AppImportedNameLabel])
	if err != nil {
		return nil, err
	}
	nonManagedApp, err := synthetic.NonManagedCamelApplicationFactory(*objOwner)
	if err != nil {
		return nil, err
	}
	targetApp := app.DeepCopy()
	targetApp.Status = v1alpha1.AppStatus{}

	deployImage := nonManagedApp.GetAppImage()
	appPhase := nonManagedApp.GetAppPhase()
	targetApp.Status.Phase = appPhase
	targetApp.Status.Image = deployImage
	pods, err := nonManagedApp.GetPods(ctx, action.client)
	if err != nil {
		return nil, err
	}
	targetApp.Status.Pods = pods
	targetApp.Status.Replicas = nonManagedApp.GetReplicas()

	return targetApp, nil
}

func lookupObject(ctx context.Context, c client.Client, kind, ns string, name string) (*ctrl.Object, error) {
	var obj ctrl.Object
	switch kind {
	case "Deployment":
		obj = &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       kind,
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		// TODO more cases
	default:
		return nil, fmt.Errorf("cannot manage Camel application of type %s", kind)
	}
	key := ctrl.ObjectKey{
		Namespace: ns,
		Name:      name,
	}
	if err := c.Get(ctx, key, obj); err != nil && k8serrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &obj, nil
}
