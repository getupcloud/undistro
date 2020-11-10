/*
Copyright 2020 Getup Cloud.
*/

package main

import (
	"flag"
	"os"
	"time"

	"github.com/getupio-undistro/undistro/internal/record"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	undistroiov1alpha1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/controllers"
	"github.com/getupio-undistro/undistro/internal/scheme"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		maxConcurrency       int
		maxConcurrencyHelm   int
		syncPeriod           time.Duration
	)
	flag.DurationVar(&syncPeriod, "sync-period", 15*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")
	flag.IntVar(&maxConcurrency, "concurrency", 10, "Number of clusters to process simultaneously")
	flag.IntVar(&maxConcurrencyHelm, "helm-concurrency", 5, "Number of helm releases to process simultaneously")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	restCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "undistro.io",
		SyncPeriod:         &syncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	ctx := ctrl.SetupSignalHandler()
	record.InitFromRecorder(mgr.GetEventRecorderFor("undistro"))
	if err = (&controllers.ClusterReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("Cluster"),
		Scheme:     mgr.GetScheme(),
		RestConfig: restCfg,
	}).SetupWithManager(ctx, mgr, concurrency(maxConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}

	if err = (&controllers.HelmReleaseReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("HelmRelease"),
		Scheme:     mgr.GetScheme(),
		RestConfig: restCfg,
	}).SetupWithManager(ctx, mgr, concurrency(maxConcurrencyHelm)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HelmRelease")
		os.Exit(1)
	}
	_, ok := os.LookupEnv("UNDISTRO_DEBUG")
	if !ok {
		if err = (&undistroiov1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Cluster")
			os.Exit(1)
		}
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func concurrency(c int) controller.Options {
	return controller.Options{MaxConcurrentReconciles: c}
}
