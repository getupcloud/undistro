/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/client/cluster/cluster"
	"github.com/getupcloud/undistro/internal/util"
	"github.com/getupcloud/undistro/status"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	clusterApi "sigs.k8s.io/cluster-api/api/v1alpha3"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	jobOwnerKey = ".metadata.controller"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

// +kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=acme.cert-manager.io,resources=*,verbs=deletecollection;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=*,verbs=deletecollection;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=*,verbs=deletecollection;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=auditregistration.k8s.io,resources=auditsinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getupcloud.com,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getupcloud.com,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=getupcloud.com,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=delete;get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;create;update;patch;delete

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cluster", req.NamespacedName)
	var cl undistrov1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cl); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't get object", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	defer status.SetObservedGenerationCluster(ctx, r.Client, &cl, cl.Generation)
	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(&cl, undistrov1.ClusterFinalizer) {
		controllerutil.AddFinalizer(&cl, undistrov1.ClusterFinalizer)
		return ctrl.Result{}, nil
	}
	undistroClient, err := uclient.New("")
	if err != nil {
		return ctrl.Result{}, err
	}
	err = util.SetVariablesFromEnvVar(ctx, util.VariablesInput{
		VariablesClient: undistroClient.GetVariables(),
		ClientSet:       r.Client,
		NamespacedName:  req.NamespacedName,
		EnvVars:         cl.Spec.InfrastructureProvider.Env,
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	action := cluster.ChooseAction(ctx, &cl)
	return ctrl.Result{}, r.run(ctx, undistroClient, action, &cl)
}

func (r *ClusterReconciler) run(
	ctx context.Context,
	uc uclient.Client,
	action undistrov1.ClusterAction,
	cl *undistrov1.Cluster) error {
	log := r.Log
	errs := errCollection{}
	upgrade := false
next:
	switch action {
	case undistrov1.InitClusterAction:
		if !upgrade {
			log.Info("running init", "name", cl.Name, "namespace", cl.Namespace)
			err := status.SetClusterPhase(ctx, r.Client, cl, undistrov1.InitializingPhase)
			if err != nil {
				log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
			}
		}
		components, err := uc.Init(uclient.InitOptions{
			Kubeconfig: uclient.Kubeconfig{
				RestConfig: r.RestConfig,
			},
			InfrastructureProviders: []string{cl.Spec.InfrastructureProvider.NameVersion()},
			TargetNamespace:         "undistro-system",
			LogUsageInstructions:    false,
		})
		if err != nil {
			log.Error(err, "couldn't init cluster", "name", cl.Name, "namespace", cl.Namespace)
			errs = append(errs, err)
			err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
			if err != nil {
				log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
			}
			break
		}
		if len(components) == 0 {
			var comp uclient.Components
			comp, err = uc.GetProviderComponents(cl.Spec.InfrastructureProvider.Name, undistrov1.InfrastructureProviderType, uclient.ComponentsOptions{
				TargetNamespace: "undistro-system",
			})
			if err != nil {
				log.Error(err, "couldn't init cluster infra", "name", cl.Name, "namespace", cl.Namespace)
				errs = append(errs, err)
				err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
				if err != nil {
					log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
				}
				break
			}
			components = append(components, comp)
		}
		ics := make([]undistrov1.InstalledComponent, len(components))
		for i, component := range components {
			preConfigFunc := component.GetPreConfigFunc()
			if preConfigFunc != nil {
				log.Info("executing pre config func", "component", component.Name())
				err = preConfigFunc(cl, uc.GetVariables())
				if err != nil {
					log.Error(err, "couldn't init cluster pre config func", "name", cl.Name, "namespace", cl.Namespace)
					errs = append(errs, err)
					err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
					if err != nil {
						log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
					}
					break
				}
			}
			ic := undistrov1.InstalledComponent{
				Name:    component.Name(),
				Version: component.Version(),
				URL:     component.URL(),
				Type:    component.Type(),
			}
			ics[i] = ic
		}
		err = status.SetClusterPhaseWithCompoments(ctx, r.Client, cl, undistrov1.InitializedPhase, ics)
		if err != nil {
			log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
		}
		action = undistrov1.ProvisionClusterAction
		if upgrade {
			action = undistrov1.UpgradeClusterAction
		}
		goto next
	case undistrov1.ProvisionClusterAction:
		log.Info("running provision", "name", cl.Name, "namespace", cl.Namespace)
		var capi unstructured.Unstructured
		tpl, err := uc.GetClusterTemplate(uclient.GetClusterTemplateOptions{
			Kubeconfig: uclient.Kubeconfig{
				RestConfig: r.RestConfig,
			},
			ClusterName:              cl.Name,
			TargetNamespace:          cl.Namespace,
			ListVariablesOnly:        false,
			KubernetesVersion:        cl.Spec.KubernetesVersion,
			ControlPlaneMachineCount: cl.Spec.ControlPlaneNode.Replicas,
			WorkerMachineCount:       cl.Spec.WorkerNode.Replicas,
		})
		if err != nil {
			log.Error(err, "couldn't get cluster template", "name", cl.Name, "namespace", cl.Namespace)
			errs = append(errs, err)
			err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
			if err != nil {
				log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
			}
			break
		}
		objs := utilresource.SortForCreate(tpl.Objs())
		for _, o := range objs {
			if o.GetKind() == "Cluster" && o.GroupVersionKind().GroupVersion().String() == clusterApi.GroupVersion.String() {
				err = ctrl.SetControllerReference(cl, &o, r.Scheme)
				if err != nil {
					log.Error(err, "couldn't set reference", "name", cl.Name, "namespace", cl.Namespace)
					errs = append(errs, err)
					err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
					if err != nil {
						log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
					}
					break
				}
				capi = o
			}
			err = r.Create(ctx, &o)
			if err != nil {
				log.Error(err, "couldn't create object", "name", o.GetName(), "namespace", o.GetNamespace())
				errs = append(errs, err)
				err = status.SetClusterPhase(ctx, r.Client, cl, undistrov1.FailedPhase)
				if err != nil {
					log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
				}
				break
			}
		}
		err = status.SetClusterPhaseWithCapi(ctx, r.Client, cl, undistrov1.ProvisioningPhase, capi)
		if err != nil {
			log.Error(err, "couldn't update status", "name", cl.Name, "namespace", cl.Namespace)
		}
	case undistrov1.StatusClusterAction:
	}
	if errs.Empty() {
		return nil
	}
	return errs
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager, opts controller.Options) error {
	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &clusterApi.Cluster{}, jobOwnerKey, func(rawObj runtime.Object) []string {
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
		WithOptions(opts).
		For(&undistrov1.Cluster{}).
		Owns(&clusterApi.Cluster{}).
		Complete(r)
}
