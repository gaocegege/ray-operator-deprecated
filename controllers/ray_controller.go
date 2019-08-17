/*
Copyright 2019 The Kubeflow community.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/composer"
	"github.com/kubeflow/ray-operator/pkg/validator"
)

const (
	ControllerName = "ray-operator"
)

// RayReconciler reconciles a Ray object
type RayReconciler struct {
	client.Client
	record.EventRecorder

	Validator validator.Interface
	Composer  composer.Interface
	Log       logr.Logger
}

// +kubebuilder:rbac:groups=ray.kubeflow.org,resources=rays,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ray.kubeflow.org,resources=rays/status,verbs=get;update;patch

func (r *RayReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("ray", req.NamespacedName)

	// Fetch the Serving instance
	instance := &rayv1.Ray{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	return r.sync(instance)
}

// SetupWithManager setups the manager and watch the deployment resource.
func (r *RayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rayv1.Ray{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &rayv1.Ray{},
			}).
		Watches(&source.Kind{Type: &corev1.Service{}},
			&handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &rayv1.Ray{},
			}).
		Complete(r)
}
