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

	"github.com/camel-tooling/camel-monitor-operator/pkg/client"
	"github.com/camel-tooling/camel-monitor-operator/pkg/platform"
	"github.com/camel-tooling/camel-monitor-operator/pkg/util/kubernetes"
	logutil "github.com/camel-tooling/camel-monitor-operator/pkg/util/log"
	consolev1 "github.com/openshift/api/console/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const consoleImageEnv = "RELATED_IMAGE_CONSOLE_PLUGIN"

var log = logutil.Log.WithName("console")

func Add(ctx context.Context, mgr manager.Manager, c client.Client) error {
	if !platform.IsCurrentOperatorGlobal() {
		log.Info("Operator is not global, skipping console deployment (ConsolePlugin is cluster-scoped)")

		return nil
	}

	hasConsoleAPI, err := kubernetes.IsAPIResourceInstalled(c, "console.openshift.io/v1", "ConsolePlugin")
	if err != nil {
		return err
	}

	if !hasConsoleAPI {
		log.Info("ConsolePlugin CRD not found, skipping console deployment")

		return nil
	}

	image := os.Getenv(consoleImageEnv)
	if image == "" {
		log.Info("RELATED_IMAGE_CONSOLE_PLUGIN not set, skipping console deployment")

		return nil
	}

	namespace := platform.GetOperatorNamespace()
	if namespace == "" {
		log.Info("Operator namespace not determined, skipping console deployment")

		return nil
	}

	r := &reconciler{
		client:    mgr.GetClient(),
		namespace: namespace,
		image:     image,
	}

	nameFilter := predicate.NewPredicateFuncs(func(obj ctrl.Object) bool {
		return obj.GetName() == pluginName
	})

	mapToConsole := handler.EnqueueRequestsFromMapFunc(
		func(_ context.Context, _ ctrl.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{
				Name:      pluginName,
				Namespace: namespace,
			}}}
		},
	)

	bootstrapCh := make(chan event.GenericEvent, 1)

	err = builder.ControllerManagedBy(mgr).
		Named("console-controller").
		Watches(&appsv1.Deployment{}, mapToConsole, builder.WithPredicates(nameFilter)).
		Watches(&corev1.Service{}, mapToConsole, builder.WithPredicates(nameFilter)).
		Watches(&corev1.ConfigMap{}, mapToConsole, builder.WithPredicates(nameFilter)).
		Watches(&consolev1.ConsolePlugin{}, mapToConsole, builder.WithPredicates(nameFilter)).
		WatchesRawSource(source.Channel(bootstrapCh, mapToConsole)).
		Complete(r)
	if err != nil {
		return err
	}

	return mgr.Add(manager.RunnableFunc(func(_ context.Context) error {
		log.Info("Triggering initial console reconciliation")

		bootstrapCh <- event.GenericEvent{
			Object: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      pluginName,
				Namespace: namespace,
			}},
		}

		return nil
	}))
}

type reconciler struct {
	client    ctrl.Client
	namespace string
	image     string
}

func (r *reconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling console resources")

	if err := r.ensureAllResources(ctx); err != nil {
		log.Error(err, "Failed console reconciliation")

		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) ensureAllResources(ctx context.Context) error {
	if err := r.ensureConfigMap(ctx); err != nil {
		return err
	}

	if err := r.ensureDeployment(ctx); err != nil {
		return err
	}

	if err := r.ensureService(ctx); err != nil {
		return err
	}

	return r.ensureConsolePlugin(ctx)
}

func (r *reconciler) ensureConfigMap(ctx context.Context) error {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      pluginName,
		Namespace: r.namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, cm, func() error {
		cm.Labels = commonLabels()
		cm.Data = map[string]string{"nginx.conf": nginxConfig()}

		return nil
	})

	return err
}

func (r *reconciler) ensureDeployment(ctx context.Context) error {
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      pluginName,
		Namespace: r.namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, deploy, func() error {
		d := deployment(r.namespace, r.image)
		deploy.Labels = d.Labels
		deploy.Spec = d.Spec

		return nil
	})

	return err
}

func (r *reconciler) ensureService(ctx context.Context) error {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      pluginName,
		Namespace: r.namespace,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, svc, func() error {
		s := service(r.namespace)
		svc.Labels = s.Labels
		svc.Annotations = s.Annotations
		svc.Spec.Ports = s.Spec.Ports
		svc.Spec.Selector = s.Spec.Selector
		svc.Spec.Type = s.Spec.Type

		return nil
	})

	return err
}

func (r *reconciler) ensureConsolePlugin(ctx context.Context) error {
	cp := &consolev1.ConsolePlugin{ObjectMeta: metav1.ObjectMeta{
		Name: pluginName,
	}}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, cp, func() error {
		p := consolePlugin(r.namespace)
		cp.Labels = p.Labels
		cp.Spec = p.Spec

		return nil
	})

	return err
}
