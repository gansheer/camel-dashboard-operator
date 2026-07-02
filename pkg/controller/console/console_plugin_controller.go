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

package console

import (
	"context"
	"os"
	"time"

	consolev1 "github.com/camel-tooling/camel-monitor-operator/pkg/apis/console/v1"
	"github.com/camel-tooling/camel-monitor-operator/pkg/client"
	"github.com/camel-tooling/camel-monitor-operator/pkg/platform"
	"github.com/camel-tooling/camel-monitor-operator/pkg/util/kubernetes"
	logutil "github.com/camel-tooling/camel-monitor-operator/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	consolePluginImageEnv = "RELATED_IMAGE_CONSOLE_PLUGIN"
	reconcileInterval     = 5 * time.Minute
)

var log = logutil.Log.WithName("console-plugin")

func Add(_ context.Context, mgr manager.Manager, c client.Client) error {
	if !platform.IsCurrentOperatorGlobal() {
		log.Info("Operator is not global, skipping console plugin deployment (ConsolePlugin is cluster-scoped)")

		return nil
	}

	hasConsolePluginAPI, err := kubernetes.IsAPIResourceInstalled(c, "console.openshift.io/v1", "ConsolePlugin")
	if err != nil {
		return err
	}

	if !hasConsolePluginAPI {
		log.Info("ConsolePlugin CRD not found, skipping console plugin deployment")

		return nil
	}

	image := os.Getenv(consolePluginImageEnv)
	if image == "" {
		log.Info("RELATED_IMAGE_CONSOLE_PLUGIN not set, skipping console plugin deployment")

		return nil
	}

	namespace := platform.GetOperatorNamespace()
	if namespace == "" {
		log.Info("Operator namespace not determined, skipping console plugin deployment")

		return nil
	}

	r := &reconciler{
		client:    c,
		reader:    mgr.GetAPIReader(),
		namespace: namespace,
		image:     image,
	}

	return mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log.Info("Starting console plugin manager")

		err := r.ensureAllResources(ctx)
		if err != nil {
			log.Error(err, "Failed initial console plugin deployment")
		}

		ticker := time.NewTicker(reconcileInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				err := r.ensureAllResources(ctx)
				if err != nil {
					log.Error(err, "Failed console plugin reconciliation")
				}
			}
		}
	}))
}

type reconciler struct {
	client    client.Client
	reader    ctrl.Reader
	namespace string
	image     string
}

func (r *reconciler) ensureAllResources(ctx context.Context) error {
	err := r.ensureConfigMap(ctx)
	if err != nil {
		return err
	}

	err = r.ensureDeployment(ctx)
	if err != nil {
		return err
	}

	err = r.ensureService(ctx)
	if err != nil {
		return err
	}

	err = r.ensureConsolePlugin(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) ensureConfigMap(ctx context.Context) error {
	desired := desiredConfigMap(r.namespace)

	existing := &corev1.ConfigMap{}
	err := r.reader.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)

	if k8serrors.IsNotFound(err) {
		log.Infof("Creating ConfigMap %s/%s", desired.Namespace, desired.Name)

		return r.client.Create(ctx, desired)
	}

	if err != nil {
		return err
	}

	existing.Labels = desired.Labels
	existing.Data = desired.Data

	return r.client.Update(ctx, existing)
}

func (r *reconciler) ensureDeployment(ctx context.Context) error {
	desired := desiredDeployment(r.namespace, r.image)

	existing := &appsv1.Deployment{}
	err := r.reader.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)

	if k8serrors.IsNotFound(err) {
		log.Infof("Creating Deployment %s/%s", desired.Namespace, desired.Name)

		return r.client.Create(ctx, desired)
	}

	if err != nil {
		return err
	}

	existing.Labels = desired.Labels
	existing.Spec = desired.Spec

	return r.client.Update(ctx, existing)
}

func (r *reconciler) ensureService(ctx context.Context) error {
	desired := desiredService(r.namespace)

	existing := &corev1.Service{}
	err := r.reader.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)

	if k8serrors.IsNotFound(err) {
		log.Infof("Creating Service %s/%s", desired.Namespace, desired.Name)

		return r.client.Create(ctx, desired)
	}

	if err != nil {
		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Spec.Ports = desired.Spec.Ports
	existing.Spec.Selector = desired.Spec.Selector
	existing.Spec.Type = desired.Spec.Type

	return r.client.Update(ctx, existing)
}

func (r *reconciler) ensureConsolePlugin(ctx context.Context) error {
	desired := desiredConsolePlugin(r.namespace)

	existing := &consolev1.ConsolePlugin{}
	err := r.reader.Get(ctx, types.NamespacedName{Name: desired.Name}, existing)

	if k8serrors.IsNotFound(err) {
		log.Infof("Creating ConsolePlugin %s", desired.Name)

		return r.client.Create(ctx, desired)
	}

	if err != nil {
		return err
	}

	existing.Labels = desired.Labels
	existing.Spec = desired.Spec

	return r.client.Update(ctx, existing)
}
