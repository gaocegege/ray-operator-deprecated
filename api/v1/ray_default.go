package v1

import (
	"github.com/kubeflow/ray-operator/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	defaultImage = "rayproject/examples"
)

var (
	defaultCmd = []string{
		"/bin/bash",
		"-c",
		"--",
	}
	defaultHeadArgs  = []string{"ray start --head --redis-port=6379 --redis-shard-ports=6380,6381 --object-manager-port=12345 --node-manager-port=12346 --node-ip-address=$RAY_NODE_IP --block"}
	defaultHeadPorts = []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "redis-primary",
			ContainerPort: 6379,
		},
		corev1.ContainerPort{
			Name:          "redis-shard-0",
			ContainerPort: 6380,
		},
		corev1.ContainerPort{
			Name:          "redis-shard-1",
			ContainerPort: 6381,
		},
		corev1.ContainerPort{
			Name:          "object-manager",
			ContainerPort: 12345,
		},
		corev1.ContainerPort{
			Name:          "node-manager",
			ContainerPort: 12346,
		},
	}

	_   webhook.Defaulter = &Ray{}
	log                   = ctrl.Log.WithName("ray-defaulter")
)

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Ray) Default() {
	log.V(1).Info("default", "name", r.Name)
	if r.Spec.Head == nil {
		r.Spec.Head = &ReplicaSpec{}
	}
	defaultHead(r.Spec.Head)
}

func defaultHead(head *ReplicaSpec) {
	if head.Replicas == nil {
		head.Replicas = int32Ptr(1)
	}
	if head.Template == nil {
		head.Template = &corev1.PodTemplateSpec{}
	}
	defaultHeadTemplate(head.Template)
}

func defaultHeadTemplate(template *corev1.PodTemplateSpec) {
	var c *v1.Container
	if !hasHeadContainer(template) {
		c = &v1.Container{
			Name: consts.ContainerRayHead,
		}
		defaultHeadContainer(c)
		template.Spec.Containers = append(template.Spec.Containers, *c)
		return
	}
	for i := range template.Spec.Containers {
		if template.Spec.Containers[i].Name == consts.ContainerRayHead {
			c = &template.Spec.Containers[i]
			defaultHeadContainer(c)
			return
		}
	}
}

func defaultHeadContainer(c *corev1.Container) {
	if c.Image == "" {
		c.Image = defaultImage
	}
	if len(c.Command) == 0 {
		c.Command = defaultCmd
	}
	if len(c.Args) == 0 {
		c.Args = defaultHeadArgs
	}
	if len(c.Ports) == 0 {
		c.Ports = defaultHeadPorts
	}
}

func hasHeadContainer(template *corev1.PodTemplateSpec) bool {
	for _, c := range template.Spec.Containers {
		if c.Name == consts.ContainerRayHead {
			return true
		}
	}
	return false
}

func int32Ptr(n int32) *int32 {
	return &n
}
