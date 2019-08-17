package composer

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/consts"
)

// DesiredHead gets the desired specification of the Head.
func (c Composer) DesiredHead(ray *rayv1.Ray) (*appsv1.Deployment, error) {
	deploymentLabels := ray.Labels
	headName := getHeadName(ray.Name)

	podLabels := getHeadPodLabels(ray.Name)
	for k, v := range ray.Labels {
		podLabels[k] = v
	}

	template := ray.Spec.Head.Template.DeepCopy()
	template.Labels = podLabels
	for i := range template.Spec.Containers {
		template.Spec.Containers[i].Env = append(template.Spec.Containers[i].Env,
			corev1.EnvVar{
				Name: consts.EnvNodeIP,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: consts.FieldPathPodIP,
					},
				},
			})
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      headName,
			Namespace: ray.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Replicas: ray.Spec.Head.Replicas,
			Template: *template,
		},
	}
	if err := controllerutil.SetControllerReference(ray, deploy, c.scheme); err != nil {
		return nil, err
	}
	return deploy, nil
}

// DesiredWorker gets the desired specificatio of the Worker.
func (c Composer) DesiredWorker(ray *rayv1.Ray) (*appsv1.Deployment, error) {
	deploymentLabels := ray.Labels
	headName := getHeadName(ray.Name)
	workerName := getWorkerName(ray.Name)

	podLabels := getWorkerPodLabels(ray.Name)
	for k, v := range ray.Labels {
		podLabels[k] = v
	}

	template := ray.Spec.Worker.Template.DeepCopy()
	template.Labels = podLabels
	for i := range template.Spec.Containers {
		template.Spec.Containers[i].Env = append(template.Spec.Containers[i].Env,
			corev1.EnvVar{
				Name: consts.EnvNodeIP,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: consts.FieldPathPodIP,
					},
				},
			},
			corev1.EnvVar{
				Name:  consts.EnvRayHeadService,
				Value: headName,
			})
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workerName,
			Namespace: ray.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Replicas: ray.Spec.Worker.Replicas,
			Template: *template,
		},
	}
	if err := controllerutil.SetControllerReference(ray, deploy, c.scheme); err != nil {
		return nil, err
	}
	return deploy, nil
}

func getHeadPodLabels(rayName string) map[string]string {
	return map[string]string{
		consts.LabelRayHead: getHeadName(rayName),
		consts.LabelRay:     rayName,
	}
}

func getWorkerPodLabels(rayName string) map[string]string {
	return map[string]string{
		consts.LabelRayWorker: getWorkerName(rayName),
		consts.LabelRay:       rayName,
	}
}

func getHeadName(rayName string) string {
	return fmt.Sprintf("%s-head", rayName)
}

func getWorkerName(rayName string) string {
	return fmt.Sprintf("%s-worker", rayName)
}
