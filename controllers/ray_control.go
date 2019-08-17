package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/consts"
)

func (r *RayReconciler) createOrUpdateService(ray *rayv1.Ray, service *corev1.Service) (*corev1.Service, error) {
	found := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		r.Log.V(1).Info("Creating Service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		if err != nil {
			r.Log.Error(err, "Failed to create the service")
			r.Event(ray, consts.EventWarning, consts.ReasonCreate,
				fmt.Sprintf("Failed to create the service %s", service.Name))
			return nil, err
		}
		r.Event(ray, consts.EventNormal, consts.ReasonCreate,
			fmt.Sprintf("Successfully create the service %s", service.Name))
		return service, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get the service")
		r.Event(ray, consts.EventWarning, consts.ReasonCreate,
			fmt.Sprintf("Failed to create the service %s", service.Name))
		return nil, err
	}

	if doServiceChanged(service, found) {
		r.Log.V(1).Info("Updating service", "namespace", service.Namespace, "name", service.Name)
		err = r.Update(context.TODO(), service)
		if err != nil {
			r.Log.Error(err, "Failed to update the service")
			r.Event(ray, consts.EventWarning, consts.ReasonUpdate,
				fmt.Sprintf("Failed to update the service %s", service.Name))
			return nil, err
		}
		r.Event(ray, consts.EventNormal, consts.ReasonUpdate,
			fmt.Sprintf("Successfully update the service %s", service.Name))
		return service, nil
	}
	return found, nil
}

func (r *RayReconciler) createOrUpdateDeployment(ray *rayv1.Ray,
	deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	// Check if the Deployment already exists.
	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		r.Log.V(1).Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			r.Log.Error(err, "Failed to create the deployment")
			r.Event(ray, consts.EventWarning, consts.ReasonCreate,
				fmt.Sprintf("Failed to create the deployment %s", deploy.Name))
			return nil, err
		}
		r.Event(ray, consts.EventNormal, consts.ReasonCreate,
			fmt.Sprintf("Successfully create the deployment %s", deploy.Name))
		return deploy, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get the deployment")
		r.Event(ray, consts.EventWarning, consts.ReasonCreate,
			fmt.Sprintf("Failed to create the deployment %s", deploy.Name))
		return nil, err
	}

	if doDeploymentChanged(deploy, found) {
		r.Log.V(1).Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), deploy)
		if err != nil {
			r.Log.Error(err, "Failed to update the deployment")
			r.Event(ray, consts.EventWarning, consts.ReasonUpdate,
				fmt.Sprintf("Failed to update the deployment %s", deploy.Name))
			return nil, err
		}
		r.Event(ray, consts.EventNormal, consts.ReasonUpdate,
			fmt.Sprintf("Successfully update the deployment %s", deploy.Name))
		return deploy, nil
	}
	return found, nil
}

// doServiceChanged checks if a serivce should be updated.
func doServiceChanged(new *corev1.Service, old *corev1.Service) bool {
	if len(new.Spec.Ports) != len(old.Spec.Ports) {
		return true
	}
	for i := range new.Spec.Ports {
		if new.Spec.Ports[i].Port != old.Spec.Ports[i].Port {
			return true
		}
	}
	return false
}

// doDeploymentChanged checks if a deployment should be updated. We will update it if the replicas
// or the resources are changed.
func doDeploymentChanged(new *appsv1.Deployment, old *appsv1.Deployment) bool {
	if *new.Spec.Replicas != *old.Spec.Replicas {
		return true
	}
	newContainers := new.Spec.Template.Spec.Containers
	oldContainers := old.Spec.Template.Spec.Containers
	for i := range newContainers {
		if !equality.Semantic.DeepEqual(newContainers[i].Resources, oldContainers[i].Resources) {
			return true
		}
	}
	return false
}
