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

	"github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	"github.com/squakez/camel-dashboard-operator/pkg/util/monitoring"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Add(ctx context.Context, mgr manager.Manager, c client.Client) error {
	return add(mgr, newReconciler(mgr, c))
}

func newReconciler(mgr manager.Manager, c client.Client) reconcile.Reconciler {
	return monitoring.NewInstrumentedReconciler(
		&reconcileApp{
			client:   c,
			reader:   mgr.GetAPIReader(),
			scheme:   mgr.GetScheme(),
			recorder: mgr.GetEventRecorderFor("camel-dashboard-app-controller"),
		},
		schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    v1alpha1.AppKind,
		},
	)
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	return builder.ControllerManagedBy(mgr).
		Named("app-controller").
		// Watch for changes to primary resource Build
		For(&v1alpha1.App{}).
		Complete(r)
}

var _ reconcile.Reconciler = &reconcileApp{}

// reconcileApp reconciles a Build object.
type reconcileApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the API server
	client client.Client
	// Non-caching client to be used whenever caching may cause race conditions,
	// like in the builds scheduling critical section
	reader   ctrl.Reader
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *reconcileApp) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
