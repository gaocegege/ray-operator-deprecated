package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Mutating struct {
}

// SetupWithManager setups the manager.
func (m *Mutating) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&rayv1.Ray{}).Name("mutating.experiment.kubeflow.org")
}
