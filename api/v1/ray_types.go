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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RaySpec defines the desired state of Ray
type RaySpec struct {
	Head   *ReplicaSpec `json:"head,omitempty"`
	Worker ReplicaSpec  `json:"worker"`
}

// ReplicaSpec is the replica specification for Head and Worker.
type ReplicaSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`

	// Describes the pod that will be created for this replica.
	Template *corev1.PodTemplateSpec `json:"template,omitempty"`
}

// RayStatus defines the observed state of Ray
type RayStatus struct {
	Head   ReplicaStatus `json:"head,omitempty"`
	Worker ReplicaStatus `json:"worker,omitempty"`
	// Conditions is an array of current observed ray conditions.
	Conditions []RayCondition `json:"conditions,omitempty"`

	// Represents time when the ray was acknowledged by the ray operator.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The generation observed by the ray operator.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents last time when the Ray was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

// ReplicaStatus is the status field for the replica.
type ReplicaStatus struct {
	// Total number of non-terminated pods targeted by this replica (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of non-terminated pods targeted by this replica that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// Total number of ready pods targeted by this replica.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this replica.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Total number of unavailable pods targeted by this replica. This is the total number of
	// pods that are still required for the replica to have 100% available capacity. They may
	// either be pods that are running but not yet available or pods that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`
}

// RayCondition is the status condition for the Ray.
type RayCondition struct {
	// Type of the condition.
	Type RayConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// RayConditionType is the type for the RayCondition.
type RayConditionType string

const (
	// RayHealth shows if the Ray is healthy.
	RayHealth RayConditionType = "Health"

	RayHeadDeploymentAvailable      RayConditionType = "RayHeadDeploymentAvailable"
	RayHeadDeploymentProgressing    RayConditionType = "RayHeadDeploymentProgressing"
	RayHeadDeploymentReplicaFailure RayConditionType = "RayHeadDeploymentReplicaFailure"

	RayWorkerDeploymentAvailable      RayConditionType = "RayWorkerDeploymentAvailable"
	RayWorkerDeploymentProgressing    RayConditionType = "RayWorkerDeploymentProgressing"
	RayWorkerDeploymentReplicaFailure RayConditionType = "RayWorkerDeploymentReplicaFailure"
)

// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status

// Ray is the Schema for the rays API
type Ray struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RaySpec   `json:"spec,omitempty"`
	Status RayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RayList contains a list of Ray
type RayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ray `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ray{}, &RayList{})
}
