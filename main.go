/*
Copyright 2020-2021 The UnDistro authors

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
	"context"
	"flag"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	appcontroller "github.com/getupio-undistro/undistro/controllers/app"
	appcontrollers "github.com/getupio-undistro/undistro/controllers/app"
	configcontroller "github.com/getupio-undistro/undistro/controllers/config"
	configcontrollers "github.com/getupio-undistro/undistro/controllers/config"
	"github.com/getupio-undistro/undistro/pkg/record"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver"
	"github.com/getupio-undistro/undistro/pkg/version"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	var metricsAddr string
	var undistroApiAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&undistroApiAddr, "undistro-api-addr", ":2020", "The address and port of the UnDistro API server")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "19acc88f.undistro.io",
		Namespace:          "",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	record.InitFromRecorder(mgr.GetEventRecorderFor("undistro-controller"))

	if err = (&configcontroller.ProviderReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Provider"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Provider")
		os.Exit(1)
	}
	if err = (&appcontroller.ClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Cluster"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err = (&appcontroller.HelmReleaseReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("HelmRelease"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HelmRelease")
		os.Exit(1)
	}
	if err = (&appcontroller.DefaultPoliciesReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DefaultPolicies"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultPolicies")
		os.Exit(1)
	}
	if err = (&configv1alpha1.Provider{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Provider")
		os.Exit(1)
	}
	if err = (&appv1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cluster")
		os.Exit(1)
	}
	if err = (&appv1alpha1.HelmRelease{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "HelmRelease")
		os.Exit(1)
	}

	if err = (&appv1alpha1.DefaultPolicies{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DefaultPolicies")
		os.Exit(1)
	}

	if err = (&configcontrollers.ProviderReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Provider")
		os.Exit(1)
	}
	if err = (&appcontrollers.ClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err = (&appcontrollers.DefaultPoliciesReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultPolicies")
		os.Exit(1)
	}
	if err = (&appcontrollers.HelmReleaseReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HelmRelease")
		os.Exit(1)
	}
	if err = (&appv1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cluster")
		os.Exit(1)
	}
	if err = (&appv1alpha1.HelmRelease{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "HelmRelease")
		os.Exit(1)
	}
	if err = (&appv1alpha1.DefaultPolicies{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DefaultPolicies")
		os.Exit(1)
	}
	if err = (&configv1alpha1.Provider{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Provider")
		os.Exit(1)
	}
	if err = (&appcontrollers.ClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err = (&appcontrollers.DefaultPoliciesReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultPolicies")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder
	cerr := make(chan error)
	done := make(chan struct{})
	go func(ctx context.Context) {
		setupLog.Info("Starting UnDistro API Server", "version", version.Get())
		server := apiserver.NewServer(cfg, os.Stdin, os.Stdout, os.Stderr)
		if err := server.GracefullyStart(ctx, undistroApiAddr); err != nil {
			cerr <- err
			return
		}
		done <- struct{}{}
	}(ctx)
	go func(ctx context.Context) {
		setupLog.Info("Starting Manager", "version", version.Get())
		if err := mgr.Start(ctx); err != nil {
			cerr <- err
			return
		}
		done <- struct{}{}
	}(ctx)

	select {
	case err := <-cerr:
		setupLog.Error(err, "failed to start UnDistro")
	case <-done:
	}
}
