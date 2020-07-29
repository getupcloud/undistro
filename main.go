/*
Copyright 2020 Getup Cloud.
*/

package main

import (
	"flag"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	bootstrapv1alpha1 "github.com/getupcloud/undistro/apis/bootstrap/v1alpha1"
	clusterv1alpha1 "github.com/getupcloud/undistro/apis/cluster/v1alpha1"
	controlplanev1alpha1 "github.com/getupcloud/undistro/apis/controlplane/v1alpha1"
	infrav1alpha1 "github.com/getupcloud/undistro/apis/infrastructure/v1alpha1"
	"github.com/getupcloud/undistro/cmd/version"
	bootstrapcontroller "github.com/getupcloud/undistro/controllers/bootstrap"
	clustercontroller "github.com/getupcloud/undistro/controllers/cluster"
	"github.com/getupcloud/undistro/controllers/cluster/remote"
	controlplanecontroller "github.com/getupcloud/undistro/controllers/controlplane"
	infracontroller "github.com/getupcloud/undistro/controllers/infrastructure"
	"github.com/getupcloud/undistro/util"
	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	"github.com/spf13/pflag"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	// flags
	metricsAddr                    string
	enableLeaderElection           bool
	leaderElectionLeaseDuration    time.Duration
	leaderElectionRenewDeadline    time.Duration
	leaderElectionRetryPeriod      time.Duration
	watchNamespace                 string
	profilerAddress                string
	clusterConcurrency             int
	machineConcurrency             int
	machineSetConcurrency          int
	machineDeploymentConcurrency   int
	machineHealthCheckConcurrency  int
	kubeadmConfigConcurrency       int
	kubeadmControlPlaneConcurrency int
	infraConcurrency               int
	syncPeriod                     time.Duration
	webhookPort                    int
	healthAddr                     string
)

func init() {
	klog.InitFlags(nil)

	_ = clientgoscheme.AddToScheme(scheme)
	_ = clusterv1alpha1.AddToScheme(scheme)
	_ = bootstrapv1alpha1.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = controlplanev1alpha1.AddToScheme(scheme)
	_ = infrav1alpha1.AddToScheme(scheme)
	_ = certmanagerv1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// InitFlags initializes the flags.
func InitFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080",
		"The address the metric endpoint binds to.")

	fs.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	fs.DurationVar(&leaderElectionLeaseDuration, "leader-election-lease-duration", 15*time.Second,
		"Interval at which non-leader candidates will wait to force acquire leadership (duration string)")

	fs.DurationVar(&leaderElectionRenewDeadline, "leader-election-renew-deadline", 10*time.Second,
		"Duration that the leading controller manager will retry refreshing leadership before giving up (duration string)")

	fs.DurationVar(&leaderElectionRetryPeriod, "leader-election-retry-period", 2*time.Second,
		"Duration the LeaderElector clients should wait between tries of actions (duration string)")

	fs.StringVar(&watchNamespace, "namespace", "",
		"Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.")

	fs.StringVar(&profilerAddress, "profiler-address", "",
		"Bind address to expose the pprof profiler (e.g. localhost:6060)")

	fs.IntVar(&clusterConcurrency, "cluster-concurrency", 10,
		"Number of clusters to process simultaneously")

	fs.IntVar(&machineConcurrency, "machine-concurrency", 10,
		"Number of machines to process simultaneously")

	fs.IntVar(&machineSetConcurrency, "machineset-concurrency", 10,
		"Number of machine sets to process simultaneously")

	fs.IntVar(&machineDeploymentConcurrency, "machinedeployment-concurrency", 10,
		"Number of machine deployments to process simultaneously")

	fs.IntVar(&machineHealthCheckConcurrency, "machinehealthcheck-concurrency", 10,
		"Number of machine health checks to process simultaneously")

	fs.DurationVar(&syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")

	fs.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook Server port, disabled by default. When enabled, the manager will only work as webhook server, no reconcilers are installed.")

	fs.StringVar(&healthAddr, "health-addr", ":9440",
		"The address the health endpoint binds to.")

	fs.IntVar(&kubeadmConfigConcurrency, "kubeadmconfig-concurrency", 10,
		"Number of kubeadm configs to process simultaneously")

	fs.IntVar(&kubeadmControlPlaneConcurrency, "kubeadmcontrolplane-concurrency", 10,
		"Number of kubeadm control planes to process simultaneously")

	fs.IntVar(&infraConcurrency, "concurrency", 10,
		"The number of docker machines to process simultaneously")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	InitFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(klogr.New())

	if profilerAddress != "" {
		klog.Infof("Profiler listening for requests at %s", profilerAddress)
		go func() {
			klog.Info(http.ListenAndServe(profilerAddress, nil))
		}()
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "undistro",
		LeaseDuration:          &leaderElectionLeaseDuration,
		RenewDeadline:          &leaderElectionRenewDeadline,
		RetryPeriod:            &leaderElectionRetryPeriod,
		Namespace:              watchNamespace,
		SyncPeriod:             &syncPeriod,
		NewClient:              util.ManagerDelegatingClientFunc,
		Port:                   webhookPort,
		HealthProbeBindAddress: healthAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupChecks(mgr)
	setupReconcilers(mgr)
	setupWebhooks(mgr)

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager", "version", version.Get().String())
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupChecks(mgr ctrl.Manager) {
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}
}

func setupReconcilers(mgr ctrl.Manager) {
	if webhookPort != 0 {
		return
	}

	// Set up a ClusterCacheTracker and ClusterCacheReconciler to provide to controllers
	// requiring a connection to a remote cluster
	tracker, err := remote.NewClusterCacheTracker(
		ctrl.Log.WithName("remote").WithName("ClusterCacheTracker"),
		mgr,
	)
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}
	if err := (&remote.ClusterCacheReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("remote").WithName("ClusterCacheReconciler"),
		Tracker: tracker,
	}).SetupWithManager(mgr, concurrency(clusterConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterCacheReconciler")
		os.Exit(1)
	}

	if err := (&clustercontroller.ClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Cluster"),
	}).SetupWithManager(mgr, concurrency(clusterConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err := (&clustercontroller.MachineReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("Machine"),
		Tracker: tracker,
	}).SetupWithManager(mgr, concurrency(machineConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Machine")
		os.Exit(1)
	}
	if err := (&clustercontroller.MachineSetReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("MachineSet"),
		Tracker: tracker,
	}).SetupWithManager(mgr, concurrency(machineSetConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MachineSet")
		os.Exit(1)
	}
	if err := (&clustercontroller.MachineDeploymentReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MachineDeployment"),
	}).SetupWithManager(mgr, concurrency(machineDeploymentConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MachineDeployment")
		os.Exit(1)
	}

	if err := (&clustercontroller.MachineHealthCheckReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("MachineHealthCheck"),
		Tracker: tracker,
	}).SetupWithManager(mgr, concurrency(machineHealthCheckConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MachineHealthCheck")
		os.Exit(1)
	}

	if err := (&bootstrapcontroller.KubeadmConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("KubeadmConfig"),
	}).SetupWithManager(mgr, concurrency(kubeadmConfigConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubeadmConfig")
		os.Exit(1)
	}

	if err := (&controlplanecontroller.KubeadmControlPlaneReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("KubeadmControlPlane"),
	}).SetupWithManager(mgr, concurrency(kubeadmControlPlaneConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubeadmControlPlane")
		os.Exit(1)
	}

	if err := (&infracontroller.DockerMachineReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DockerMachine"),
	}).SetupWithManager(mgr, concurrency(infraConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "reconciler")
		os.Exit(1)
	}

	if err := (&infracontroller.DockerClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DockerCluster"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DockerCluster")
		os.Exit(1)
	}
}

func setupWebhooks(mgr ctrl.Manager) {
	if webhookPort == 0 {
		return
	}

	if err := (&clusterv1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cluster")
		os.Exit(1)
	}

	if err := (&clusterv1alpha1.Machine{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Machine")
		os.Exit(1)
	}

	if err := (&clusterv1alpha1.MachineSet{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MachineSet")
		os.Exit(1)
	}

	if err := (&clusterv1alpha1.MachineDeployment{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MachineDeployment")
		os.Exit(1)
	}

	if err := (&clusterv1alpha1.MachineHealthCheck{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MachineHealthCheck")
		os.Exit(1)
	}

	if err := (&bootstrapv1alpha1.KubeadmConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KubeadmConfig")
		os.Exit(1)
	}

	if err := (&bootstrapv1alpha1.KubeadmConfigList{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KubeadmConfigList")
		os.Exit(1)
	}

	if err := (&bootstrapv1alpha1.KubeadmConfigTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KubeadmConfigTemplate")
		os.Exit(1)
	}

	if err := (&bootstrapv1alpha1.KubeadmConfigTemplateList{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KubeadmConfigTemplateList")
		os.Exit(1)
	}
	if err := (&controlplanev1alpha1.KubeadmControlPlane{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KubeadmControlPlane")
		os.Exit(1)
	}
	if err := (&infrav1alpha1.DockerMachineTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DockerMachineTemplate")
		os.Exit(1)
	}
}

func concurrency(c int) controller.Options {
	return controller.Options{MaxConcurrentReconciles: c}
}
