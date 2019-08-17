package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/consts"
)

func (r *RayReconciler) updateStatus(ray *rayv1.Ray,
	head *appsv1.Deployment, worker *appsv1.Deployment) error {
	old := ray.Status.DeepCopy()
	status := &ray.Status

	now := metav1.Now()
	if status.StartTime == nil {
		status.StartTime = &now
	}
	status.LastReconcileTime = &now
	if ray.Generation > status.ObservedGeneration {
		status.ObservedGeneration = ray.Generation
	}

	// shouldActive is the number of the components should be active.
	shouldActive := 2
	// activeCounter is the number of components which are actually active.
	activeCounter := 0
	// running is the number of components which are actually running or pending.
	running := 0

	// Set the worker status.
	status.Worker.Replicas = worker.Status.Replicas
	status.Worker.ReadyReplicas = worker.Status.ReadyReplicas
	status.Worker.AvailableReplicas = worker.Status.AvailableReplicas
	status.Worker.UnavailableReplicas = worker.Status.UnavailableReplicas
	status.Worker.UpdatedReplicas = worker.Status.UpdatedReplicas

	// Set the head status.
	status.Head.Replicas = head.Status.Replicas
	status.Head.ReadyReplicas = head.Status.ReadyReplicas
	status.Head.AvailableReplicas = head.Status.AvailableReplicas
	status.Head.UnavailableReplicas = head.Status.UnavailableReplicas
	status.Head.UpdatedReplicas = head.Status.UpdatedReplicas

	// Get the pods belong to the deployment.
	pods := &corev1.PodList{}
	if err := r.List(context.TODO(), pods, client.MatchingLabels(map[string]string{
		consts.LabelRayWorker: worker.Name,
	})); err != nil {
		return err
	}

	if hasDeploymentAvailable(worker) {
		// Check if the deployment is available.
		if isDeploymentAvailable(worker) {
			// If the deployment is active, we set the condition to serving status
			// and mark it active.
			activeCounter++
			createOrUpdateCondition(status, rayv1.RayWorkerDeploymentAvailable,
				corev1.ConditionTrue)
		} else {
			createOrUpdateCondition(status, rayv1.RayWorkerDeploymentAvailable,
				corev1.ConditionFalse)
			// If the deployment is not active but all the pods owned by the deployment is pending or running, we mark the deployment running.
			// This is a workaround to avoid http://jira.caicloud.xyz/browse/CLV-545.
			if allPodsArePendingOrRunning(pods) {
				running++
			}
		}
	} else {
		// If the available condition is not found, we mark the deployment running.
		running++
	}
	syncDeploymentConditions(status, worker.Status.Conditions, consts.LabelRayWorker)

	// Get the pods belong to the deployment.
	pods = &corev1.PodList{}
	if err := r.List(context.TODO(), pods, client.MatchingLabels(map[string]string{
		consts.LabelRayHead: head.Name,
	})); err != nil {
		return err
	}
	if hasDeploymentAvailable(head) {
		// Check if the deployment is available.
		if isDeploymentAvailable(head) {
			// If the deployment is active, we set the condition to serving status
			// and mark it active.
			activeCounter++
			createOrUpdateCondition(status, rayv1.RayHeadDeploymentAvailable,
				corev1.ConditionTrue)
		} else {
			createOrUpdateCondition(status, rayv1.RayHeadDeploymentAvailable,
				corev1.ConditionFalse)
			// If the deployment is not active but all the pods owned by the deployment is pending or running, we mark the deployment running.
			// This is a workaround to avoid http://jira.caicloud.xyz/browse/CLV-545.
			if allPodsArePendingOrRunning(pods) {
				running++
			}
		}
	} else {
		// If the available condition is not found, we mark the deployment running.
		running++
	}
	syncDeploymentConditions(status, head.Status.Conditions, consts.LabelRayHead)

	// If all resources work well, set the healthy.
	if shouldActive == activeCounter {
		createOrUpdateCondition(status, rayv1.RayHealth,
			corev1.ConditionTrue)
	} else if shouldActive == activeCounter+running {
		createOrUpdateCondition(status, rayv1.RayHealth,
			corev1.ConditionUnknown)
	} else {
		createOrUpdateCondition(status, rayv1.RayHealth,
			corev1.ConditionFalse)
	}

	// Update the found object and write the result back if there are any changes.
	if !equality.Semantic.DeepEqual(status, old) {
		r.Log.V(1).Info("Updating Ray status", "namespace", ray.Namespace,
			"name", ray.Name,
			"status", ray.Status)
		if err := r.Status().Update(context.TODO(), ray); err != nil {
			return err
		}
	}
	return nil
}

// syncDeploymentConditions syncs deployment conditions to serving conditions.
func syncDeploymentConditions(status *rayv1.RayStatus,
	conditions []appsv1.DeploymentCondition,
	label string) {
	switch label {
	case consts.LabelRayHead:
		found := false
		for _, condition := range conditions {
			switch condition.Type {
			case appsv1.DeploymentProgressing:
				createOrUpdateConditionWithReason(status, rayv1.RayHeadDeploymentProgressing,
					condition.Status, condition.Reason, condition.Message)
			case appsv1.DeploymentReplicaFailure:
				found = true
				createOrUpdateConditionWithReason(status, rayv1.RayHeadDeploymentReplicaFailure,
					condition.Status, condition.Reason, condition.Message)
			}
		}
		if !found {
			setConditionUnknown(status, rayv1.RayHeadDeploymentReplicaFailure)
		}
	case consts.LabelRayWorker:
		found := false
		for _, condition := range conditions {
			switch condition.Type {
			case appsv1.DeploymentProgressing:
				createOrUpdateConditionWithReason(status, rayv1.RayWorkerDeploymentProgressing,
					condition.Status, condition.Reason, condition.Message)
			case appsv1.DeploymentReplicaFailure:
				found = true
				createOrUpdateConditionWithReason(status, rayv1.RayWorkerDeploymentReplicaFailure,
					condition.Status, condition.Reason, condition.Message)
			}
		}
		if !found {
			setConditionUnknown(status, rayv1.RayWorkerDeploymentReplicaFailure)
		}
	}
}

func setConditionUnknown(status *rayv1.RayStatus,
	conditionType rayv1.RayConditionType) {
	createOrUpdateCondition(status, conditionType, corev1.ConditionUnknown)
}

func createOrUpdateConditionWithReason(status *rayv1.RayStatus,
	conditionType rayv1.RayConditionType,
	boolVal corev1.ConditionStatus, reason, msg string) {
	if !containConditionType(status, conditionType) {
		status.Conditions = append(status.Conditions, newCondition(conditionType,
			boolVal, reason, msg))
	} else {
		for i := range status.Conditions {
			if status.Conditions[i].Type == conditionType {
				if status.Conditions[i].Status != boolVal {
					status.Conditions[i].LastTransitionTime = metav1.Now()
				}
				status.Conditions[i].Status = boolVal
				status.Conditions[i].LastUpdateTime = metav1.Now()
				status.Conditions[i].Reason = reason
				status.Conditions[i].Message = msg
			}
		}
	}
}

func createOrUpdateCondition(status *rayv1.RayStatus,
	conditionType rayv1.RayConditionType,
	boolVal corev1.ConditionStatus) {
	createOrUpdateConditionWithReason(status, conditionType, boolVal, "", "")
}

func containConditionType(status *rayv1.RayStatus,
	conditionType rayv1.RayConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == conditionType {
			return true
		}
	}
	return false
}

func newCondition(conditionType rayv1.RayConditionType,
	boolVal corev1.ConditionStatus,
	reason, message string) rayv1.RayCondition {
	return rayv1.RayCondition{
		Type:               conditionType,
		Status:             boolVal,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

func isDeploymentAvailable(deploy *appsv1.Deployment) bool {
	for _, condition := range deploy.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable &&
			condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func hasDeploymentAvailable(deploy *appsv1.Deployment) bool {
	for _, condition := range deploy.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable {
			return true
		}
	}
	return false
}

func allPodsArePendingOrRunning(pods *corev1.PodList) bool {
	for _, p := range pods.Items {
		if p.Status.Phase != corev1.PodPending && p.Status.Phase != corev1.PodRunning {
			return false
		}
	}
	return true
}
