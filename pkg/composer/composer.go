package composer

import (
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
)

const (
	ComposerName = "ray-operator-composer"
)

// Interface composes the desired specification for Head and Worker.
type Interface interface {
	DesiredHead(ray *rayv1.Ray) (*appsv1.Deployment, error)
	DesiredWorker(ray *rayv1.Ray) (*appsv1.Deployment, error)
	DesiredHeadService(ray *rayv1.Ray) (*corev1.Service, error)
}

// Composer is the default implementation for the Interface.
type Composer struct {
	record.EventRecorder
	Log    logr.Logger
	scheme *runtime.Scheme
}

// New returns a new composer.
func New(recorder record.EventRecorder, log logr.Logger, scheme *runtime.Scheme) Interface {
	return &Composer{
		EventRecorder: recorder,
		Log:           log,
		scheme:        scheme,
	}
}
