package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
)

// sync handles all requests.
func (r *RayReconciler) sync(ray *rayv1.Ray) (ctrl.Result, error) {
	r.Log.V(1).Info("Sync the object Ray", "namespace", ray.Namespace, "instance", ray.Name)
	defer r.Log.V(1).Info("Finished syncing Ray", "namespace", ray.Namespace, "instance", ray.Name)

	// TODO(gaocegege): We should do the validation in Validation Webhook.
	if err := r.Validator.ValidateRay(ray); err != nil {
		// Do not requeue the requests since we cannot deal with it.
		return ctrl.Result{}, nil
	}

	desiredHeadService, err := r.Composer.DesiredHeadService(ray)
	if err != nil {
		// Do not requeue the requests since we cannot deal with it.
		return ctrl.Result{}, nil
	}
	// TODO(gaocegege): Update status according to the actualService.
	if _, err := r.createOrUpdateService(ray, desiredHeadService); err != nil {
		return ctrl.Result{
			Requeue: true,
		}, nil
	}

	desiredHead, err := r.Composer.DesiredHead(ray)
	if err != nil {
		// Do not requeue the requests since we cannot deal with it.
		return ctrl.Result{}, nil
	}

	actualHead, err := r.createOrUpdateDeployment(ray, desiredHead)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, nil
	}

	desiredWorker, err := r.Composer.DesiredWorker(ray)
	if err != nil {
		// Do not requeue the requests since we cannot deal with it.
		return ctrl.Result{}, nil
	}

	actualWorker, err := r.createOrUpdateDeployment(ray, desiredWorker)
	if err != nil {
		return ctrl.Result{
			Requeue: true,
		}, nil
	}

	// Update Serving status according to the deployment, pvc and hpa.
	if err := r.updateStatus(ray, actualHead, actualWorker); err != nil {
		r.Log.Error(err, "Failed to update the status for ray", "instance", ray.Name)
		return reconcile.Result{
			Requeue: true,
		}, nil
	}
	return ctrl.Result{}, nil
}
