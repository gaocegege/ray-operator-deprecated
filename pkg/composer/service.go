package composer

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/consts"
)

func (c Composer) DesiredHeadService(ray *rayv1.Ray) (*corev1.Service, error) {
	serviceLabels := ray.Labels

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getHeadName(ray.Name),
			Namespace: ray.Namespace,
			Labels:    serviceLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: getHeadPodLabels(ray.Name),
		},
	}

	for _, c := range ray.Spec.Head.Template.Spec.Containers {
		if c.Name == consts.ContainerRayHead {
			for i, p := range c.Ports {
				name := p.Name
				if name == "" {
					name = fmt.Sprintf("copy-from-%d", i)
				}
				service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
					Name:       name,
					Port:       p.ContainerPort,
					TargetPort: intstr.FromInt(int(p.ContainerPort)),
				})
			}
		}
	}
	if err := controllerutil.SetControllerReference(ray, service, c.scheme); err != nil {
		return nil, err
	}
	return service, nil
}
