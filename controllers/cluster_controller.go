/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"
	"time"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/internal/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clusterApi "sigs.k8s.io/cluster-api/api/v1alpha3"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheduledResult = ctrl.Result{RequeueAfter: 5 * time.Second}
	jobOwnerKey     = ".metadata.controller"
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
	if cluster.Status.Phase == undistrov1.NewPhase {
		log.Info("ensure mangement cluster is initialized and updated", "name", req.NamespacedName)
		if err = r.init(ctx, &cluster, undistroClient); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't initialize or update the mangement cluster", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		if err = r.Status().Update(ctx, &cluster); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return scheduledResult, nil
	}
	if cluster.Status.Phase == undistrov1.InitializedPhase {
		log.Info("generanting cluster-api configuration", "name", req.NamespacedName)
		if err = r.config(ctx, &cluster, undistroClient); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't initialize or update the mangement cluster", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		if err = r.Status().Update(ctx, &cluster); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return scheduledResult, nil
	}
	if cluster.Status.Phase == undistrov1.ProvisioningPhase {
		res, err := r.provisioning(ctx, &cluster, undistroClient)
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't provision the cluster", "name", req.NamespacedName)
			return res, err
		}
		return res, nil
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) init(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) error {
	log := r.Log
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
		preConfigFunc := component.GetPreConfigFunc()
		if preConfigFunc != nil {
			log.Info("executing pre config func", "component", component.Name())
			err = preConfigFunc(cl, c.GetVariables())
			if err != nil {
				return err
			}
		}
		ic := undistrov1.InstalledComponent{
			Name:    component.Name(),
			Version: component.Version(),
			URL:     component.URL(),
			Type:    component.Type(),
		}
		cl.Status.InstalledComponents[i] = ic
	}
	cl.Status.Phase = undistrov1.InitializedPhase
	return nil
}

func (r *ClusterReconciler) config(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) error {
	tpl, err := c.GetClusterTemplate(uclient.GetClusterTemplateOptions{
		ClusterName:              cl.Name,
		TargetNamespace:          cl.Namespace,
		ListVariablesOnly:        false,
		KubernetesVersion:        cl.Spec.KubernetesVersion,
		ControlPlaneMachineCount: cl.Spec.ControlPlaneNode.Replicas,
		WorkerMachineCount:       cl.Spec.WorkerNode.Replicas,
	})
	if err != nil {
		return err
	}
	objs := utilresource.SortForCreate(tpl.Objs())
	for _, o := range objs {
		err = ctrl.SetControllerReference(cl, &o, r.Scheme)
		if err != nil {
			return errors.Errorf("couldn't set reference: %v", err)
		}
		err = r.Create(ctx, &o)
		if err != nil {
			return err
		}
	}
	cl.Status.Phase = undistrov1.ProvisioningPhase
	return nil
}

func (r *ClusterReconciler) provisioning(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) (ctrl.Result, error) {
	log := r.Log
	log.Info("provisioning cluster", "name", cl.Name, "namespace", cl.Namespace)
	var clusterList clusterApi.ClusterList
	if err := r.List(ctx, &clusterList, client.InNamespace(cl.Namespace), client.MatchingFields{jobOwnerKey: cl.Name}); err != nil {
		return scheduledResult, err
	}
	if len(clusterList.Items) != 1 {
		return scheduledResult, errors.Errorf("has more than one Cluster API Cluster owned by %s/%s", cl.Namespace, cl.Name)
	}
	capi := &clusterList.Items[0]
	if undistrov1.ClusterPhase(capi.Status.Phase) == undistrov1.FailedPhase {
		cl.Status.Phase = undistrov1.FailedPhase
		if err := r.Status().Update(ctx, cl); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
			return scheduledResult, err
		}
		return scheduledResult, nil
	}
	if capi.Status.ControlPlaneInitialized && !capi.Status.ControlPlaneReady {
		if err := r.installCNI(ctx, cl, c); err != nil {
			return scheduledResult, err
		}
	}
	if capi.Status.ControlPlaneReady && capi.Status.InfrastructureReady && undistrov1.ClusterPhase(capi.Status.Phase) == undistrov1.ProvisionedPhase {
		cl.Status.Phase = undistrov1.ProvisionedPhase
		cl.Status.Ready = true
		if err := r.Status().Update(ctx, cl); client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
			return scheduledResult, err
		}
		return ctrl.Result{}, nil
	}
	return scheduledResult, nil
}

func (r *ClusterReconciler) installCNI(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) error {
	log := r.Log
	cniAddr := cl.GetCNITemplateURL()
	log.Info("getting CNI", "name", cl.Spec.CniName, "URL", cniAddr)
	tpl, err := c.GetClusterTemplate(uclient.GetClusterTemplateOptions{
		ClusterName:     cl.Name,
		TargetNamespace: cl.Namespace,
		URLSource: &uclient.URLSourceOptions{
			URL: cniAddr,
		},
	})
	if err != nil {
		return err
	}
	clKubeconfig, err := c.GetKubeconfig(uclient.GetKubeconfigOptions{
		WorkloadClusterName: cl.Name,
		Namespace:           cl.Namespace,
	})
	if err != nil {
		return err
	}
	cfg, err := clientcmd.NewClientConfigFromBytes([]byte(clKubeconfig))
	if err != nil {
		return err
	}
	workloadCfg, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	workloadClient, err := client.New(workloadCfg, client.Options{Scheme: r.Scheme})
	if err != nil {
		return err
	}
	objs := utilresource.SortForCreate(tpl.Objs())
	for _, o := range objs {
		err = workloadClient.Create(ctx, &o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&clusterApi.Cluster{}, jobOwnerKey, func(rawObj runtime.Object) []string {
		cluster := rawObj.(*clusterApi.Cluster)
		owner := metav1.GetControllerOf(cluster)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != undistrov1.GroupVersion.String() || owner.Kind != "Cluster" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&undistrov1.Cluster{}).
		Owns(&clusterApi.Cluster{}).
		Complete(r)
}
