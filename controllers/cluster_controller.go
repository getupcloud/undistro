/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/internal/util"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=getupcloud.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getupcloud.com,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cluster", req.NamespacedName)
	var cluster undistrov1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); client.IgnoreNotFound(err) != nil {
		log.Error(err, "couldn't get object", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	undistroClient, err := uclient.New("")
	if err != nil {
		return ctrl.Result{}, err
	}
	err = util.SetVariablesFromEnvVar(ctx, util.VariablesInput{
		VariablesClient: undistroClient.GetVariables(),
		ClientSet:       r.Client,
		NamespacedName:  req.NamespacedName,
		EnvVars:         cluster.Spec.InfrastructureProvider.Env,
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster.Status.Phase == "" {
		log.Info("ensure mangement cluster is initialized and updated")
		if err = r.init(&cluster, undistroClient); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't initialize or update the mangement cluster")
			return ctrl.Result{}, err
		}
		if err = r.Status().Update(ctx, &cluster); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) init(cl *undistrov1.Cluster, c uclient.Client) error {
	components, err := c.Init(uclient.InitOptions{
		InfrastructureProviders: []string{cl.Spec.InfrastructureProvider.NameVersion()},
		TargetNamespace:         "undistro-system",
		LogUsageInstructions:    false,
	})
	if err != nil {
		return err
	}
	cl.Status.InstalledComponents = make([]undistrov1.InstalledComponent, len(components))
	for i, component := range components {
		ic := undistrov1.InstalledComponent{
			Name:    component.Name(),
			Version: component.Version(),
			URL:     component.URL(),
			Type:    component.Type(),
		}
		cl.Status.InstalledComponents[i] = ic
	}
	cl.Status.Phase = "initialized"
	return nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&undistrov1.Cluster{}).
		Complete(r)
}
