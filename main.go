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

package main

import (
	"flag"
	"os"

	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	rayv1 "github.com/kubeflow/ray-operator/api/v1"
	"github.com/kubeflow/ray-operator/controllers"
	"github.com/kubeflow/ray-operator/pkg/composer"
	"github.com/kubeflow/ray-operator/pkg/validator"
)

var (
	k8sScheme = scheme.Scheme
	setupLog  = ctrl.Log.WithName("setup")
)

func init() {
	_ = rayv1.AddToScheme(k8sScheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             k8sScheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	composer := composer.New(
		mgr.GetEventRecorderFor(composer.ComposerName),
		ctrl.Log.WithName(composer.ComposerName),
		mgr.GetScheme(),
	)
	validator := validator.New(
		mgr.GetEventRecorderFor(validator.ValidatorName),
		ctrl.Log.WithName(validator.ValidatorName),
	)

	if err := (&controllers.RayReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor(controllers.ControllerName),
		Composer:      composer,
		Validator:     validator,
		Log:           ctrl.Log.WithName(controllers.ControllerName).WithName("Ray"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ray")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
