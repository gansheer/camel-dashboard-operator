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
	"reflect"

	v1alpha1 "github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	"github.com/squakez/camel-dashboard-operator/pkg/platform"
	"github.com/squakez/camel-dashboard-operator/pkg/util/kubernetes"
	"github.com/squakez/camel-dashboard-operator/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgocache "k8s.io/client-go/tools/cache"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	controller = true
)

// ManageSyntheticCamelApps is the controller for synthetic Camel Applications. Consider that the lifecycle of the objects are driven
// by the way we are monitoring them. Since we're filtering by some label in the cached client, you must consider an add, update or delete
// accordingly, ie, when the user label the resource, then it is considered as an add, when it removes the label, it is considered as a delete.
func ManageSyntheticCamelApps(ctx context.Context, c client.Client, cache cache.Cache) error {
	informers, err := getInformers(ctx, c, cache)
	if err != nil {
		return err
	}
	for _, informer := range informers {
		_, err := informer.AddEventHandler(clientgocache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ctrlObj, ok := obj.(ctrl.Object)
				if !ok {
					log.Error(fmt.Errorf("type assertion failed: %v", obj), "Failed to retrieve Object on add event")
					return
				}

				onAdd(ctx, c, ctrlObj)
			},
			DeleteFunc: func(obj interface{}) {
				ctrlObj, ok := obj.(ctrl.Object)
				if !ok {
					log.Errorf(fmt.Errorf("type assertion failed: %v", obj), "Failed to retrieve Object on delete event")
					return
				}

				onDelete(ctx, c, ctrlObj)
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: refactor this code and remove lint exclusion
//
//nolint:nestif
func onAdd(ctx context.Context, c client.Client, ctrlObj ctrl.Object) {
	log.Infof("Detected a new %s resource named %s in namespace %s",
		ctrlObj.GetObjectKind().GroupVersionKind().Kind, ctrlObj.GetName(), ctrlObj.GetNamespace())
	appName := ctrlObj.GetLabels()[v1alpha1.AppLabel]
	app, err := getSyntheticCamelApp(ctx, c, ctrlObj.GetNamespace(), appName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			adapter, err := nonManagedCamelApplicationFactory(ctrlObj)
			if err != nil {
				log.Errorf(err, "Some error happened while creating a Camel application adapter for %s", appName)
			}
			if err = createSyntheticCamelApp(ctx, c, adapter.CamelApp()); err != nil {
				log.Errorf(err, "Some error happened while creating a synthetic Camel Application %s", appName)
			}
			log.Infof("Created a synthetic Camel Application %s after %s resource object", app.GetName(), ctrlObj.GetName())
		} else {
			log.Errorf(err, "Some error happened while loading a synthetic Camel Application %s", appName)
		}
	} else {
		log.Infof("Synthetic Camel Application %s is in phase %s. Skipping.", appName, app.Status.Phase)
	}
}

func onDelete(ctx context.Context, c client.Client, ctrlObj ctrl.Object) {
	appName := ctrlObj.GetLabels()[v1alpha1.AppLabel]
	// Importing label removed
	if err := deleteSyntheticCamelApp(ctx, c, ctrlObj.GetNamespace(), appName); err != nil {
		log.Errorf(err, "Some error happened while deleting a synthetic Camel Application %s", appName)
	}
	log.Infof("Deleted synthetic Camel Application %s", appName)
}

func getInformers(ctx context.Context, cl client.Client, c cache.Cache) ([]cache.Informer, error) {
	deploy, err := c.GetInformer(ctx, &appsv1.Deployment{})
	if err != nil {
		return nil, err
	}
	informers := []cache.Informer{deploy}
	// Watch for the CronJob conditionally
	if ok, err := kubernetes.IsAPIResourceInstalled(cl, batchv1.SchemeGroupVersion.String(), reflect.TypeOf(batchv1.CronJob{}).Name()); ok && err == nil {
		cron, err := c.GetInformer(ctx, &batchv1.CronJob{})
		if err != nil {
			return nil, err
		}
		informers = append(informers, cron)
	}
	// Watch for the Knative Services conditionally
	if ok, err := kubernetes.IsAPIResourceInstalled(cl, servingv1.SchemeGroupVersion.String(), reflect.TypeOf(servingv1.Service{}).Name()); ok && err == nil {
		if ok, err := kubernetes.CheckPermission(ctx, cl, serving.GroupName, "services", platform.GetOperatorWatchNamespace(), "", "watch"); ok && err == nil {
			ksvc, err := c.GetInformer(ctx, &servingv1.Service{})
			if err != nil {
				return nil, err
			}
			informers = append(informers, ksvc)
		}
	}

	return informers, nil
}

func getSyntheticCamelApp(ctx context.Context, c client.Client, namespace, name string) (*v1alpha1.App, error) {
	app := v1alpha1.NewApp(namespace, name)
	err := c.Get(ctx, ctrl.ObjectKeyFromObject(&app), &app)
	return &app, err
}

func createSyntheticCamelApp(ctx context.Context, c client.Client, it *v1alpha1.App) error {
	return c.Create(ctx, it, ctrl.FieldOwner("camel-dashboard-operator"))
}

func deleteSyntheticCamelApp(ctx context.Context, c client.Client, namespace, name string) error {
	// As the Integration label was removed, we don't know which is the Synthetic Camel Application to remove
	app := v1alpha1.NewApp(namespace, name)
	return c.Delete(ctx, &app)
}

// nonManagedCamelApplicationAdapter represents a Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelApplicationAdapter interface {
	// CamelApp return a CamelApp resource fed by the Camel application adapter.
	CamelApp() *v1alpha1.App
}

func nonManagedCamelApplicationFactory(obj ctrl.Object) (nonManagedCamelApplicationAdapter, error) {
	deploy, ok := obj.(*appsv1.Deployment)
	if ok {
		return &nonManagedCamelDeployment{deploy: deploy}, nil
	}
	cronjob, ok := obj.(*batchv1.CronJob)
	if ok {
		return &NonManagedCamelCronjob{cron: cronjob}, nil
	}
	ksvc, ok := obj.(*servingv1.Service)
	if ok {
		return &NonManagedCamelKnativeService{ksvc: ksvc}, nil
	}
	return nil, fmt.Errorf("unsupported %s object kind", obj.GetName())
}

// NonManagedCamelDeployment represents a regular Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelDeployment struct {
	deploy *appsv1.Deployment
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelDeployment) CamelApp() *v1alpha1.App {
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
	return &newApp
}

// NonManagedCamelCronjob represents a cron Camel application built and deployed outside the operator lifecycle.
type NonManagedCamelCronjob struct {
	cron *batchv1.CronJob
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *NonManagedCamelCronjob) CamelApp() *v1alpha1.App {
	newApp := v1alpha1.NewApp(app.cron.Namespace, app.cron.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.cron.Name,
		v1alpha1.AppImportedKindLabel: "CronJob",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       app.cron.Name,
			UID:        app.cron.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)
	return &newApp
}

// NonManagedCamelKnativeService represents a Knative Service based Camel application built and deployed outside the operator lifecycle.
type NonManagedCamelKnativeService struct {
	ksvc *servingv1.Service
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *NonManagedCamelKnativeService) CamelApp() *v1alpha1.App {
	newApp := v1alpha1.NewApp(app.ksvc.Namespace, app.ksvc.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.ksvc.Name,
		v1alpha1.AppImportedKindLabel: "KnativeService",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: servingv1.SchemeGroupVersion.String(),
			Kind:       "Service",
			Name:       app.ksvc.Name,
			UID:        app.ksvc.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)
	return &newApp
}
