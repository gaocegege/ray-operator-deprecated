package validator

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/pkg/consts"
)

const (
	ValidatorName = "ray-operator-validator"
)

// Interface validates the Ray specification.
type Interface interface {
	ValidateRay(ray *rayv1.Ray) error
}

// Validator is the default implementation for the Interface.
type Validator struct {
	record.EventRecorder
	Log logr.Logger
}

// New returns a new Validator.
func New(recorder record.EventRecorder, log logr.Logger) Interface {
	return &Validator{
		EventRecorder: recorder,
		Log:           log,
	}
}

// ValidateRay validates a Ray specification.
func (v Validator) ValidateRay(ray *rayv1.Ray) error {
	v.Event(ray, consts.EventWarning, consts.ReasonValidationFailed, "Not Implemented")
	v.Log.V(1).Info("Validation is not implemented, return nil")
	return nil
}
