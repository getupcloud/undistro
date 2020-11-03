/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	uclient "github.com/getupio-undistro/undistro/client"
	"github.com/getupio-undistro/undistro/client/config"
	"github.com/getupio-undistro/undistro/internal/patch"
	"github.com/getupio-undistro/undistro/internal/record"
	"github.com/getupio-undistro/undistro/internal/template"
	"github.com/getupio-undistro/undistro/internal/util"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	clusterApi "sigs.k8s.io/cluster-api/api/v1alpha3"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	"sigs.k8s.io/cluster-api/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	jobOwnerKey = ".metadata.controller"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	RestConfig   *rest.Config
	EventWatcher chan record.WatcherResource
}

// +kubebuilder:rbac:urls=/metrics,verbs=get;
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
// +kubebuilder:rbac:groups=undistro.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=undistro.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=undistro.io,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=delete;get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("cluster", req.NamespacedName)
	var cluster undistrov1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't get object", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(&cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		err = patch.ControllerObject(ctx, patchHelper, &cluster, err)
	}()
	if !controllerutil.ContainsFinalizer(&cluster, undistrov1.Finalizer) {
		controllerutil.AddFinalizer(&cluster, undistrov1.Finalizer)
		return ctrl.Result{}, nil
	}
	if !cluster.DeletionTimestamp.IsZero() {
		log.Info("removing cluster", "name", req.NamespacedName)
		if err := r.delete(ctx, &cluster); err != nil {
			return ctrl.Result{}, err
		}
		if cluster.Status.Phase == undistrov1.DeletingPhase {
			return ctrl.Result{Requeue: true}, nil
		}
		controllerutil.RemoveFinalizer(&cluster, undistrov1.Finalizer)
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
		EnvVars:         cluster.Spec.InfrastructureProvider.Env,
	})
	if err != nil {
		record.Warn(&cluster, "FailedPrepareEnvironment", "failed to prepare environment to Cluster API")
		return ctrl.Result{}, err
	}
	opts, err := config.SetupTemplates(&cluster, undistroClient.GetVariables())
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster.Status.Phase == undistrov1.NewPhase {
		log.Info("ensure mangement cluster is initialized and updated", "name", req.NamespacedName)
		if err = r.init(ctx, &cluster, undistroClient); err != nil {
			if client.IgnoreNotFound(err) != nil {
				record.Warn(&cluster, "FailedInitialize", "failed to initalize Cluster API")
				log.Error(err, "couldn't initialize or update the mangement cluster", "name", req.NamespacedName)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		record.Event(&cluster, "ClusterInitialized", "Cluster API is successfully initialized")
		return ctrl.Result{Requeue: true}, nil
	}
	if cluster.Status.Phase == undistrov1.InitializedPhase {
		log.Info("generanting cluster-api configuration", "name", req.NamespacedName)
		if err = r.config(ctx, &cluster, opts, undistroClient); err != nil {
			if client.IgnoreNotFound(err) != nil {
				record.Warn(&cluster, "FailedInstallConfig", "failed to install Cluster API objects")
				log.Error(err, "couldn't install cluster config", "name", req.NamespacedName)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		record.Event(&cluster, "ClusterInstalled", "Cluster API is successfully configured")
		return ctrl.Result{}, nil
	}
	if cluster.Status.Phase == undistrov1.ProvisioningPhase {
		res, err := r.provisioning(ctx, &cluster, undistroClient)
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't provision the cluster", "name", req.NamespacedName)
			return res, err
		}
		return res, nil
	}
	if r.hasDiff(ctx, &cluster) {
		return r.upgrade(ctx, &cluster, undistroClient, opts)
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) hasDiff(ctx context.Context, cl *undistrov1.Cluster) bool {
	diffCP := cmp.Diff(cl.Spec.ControlPlaneNode, cl.Status.ControlPlaneNode)
	diffW := cmp.Diff(cl.Spec.WorkerNodes, cl.Status.WorkerNodes)
	diffBastion := cmp.Diff(cl.Spec.Bastion, cl.Status.BastionConfig)
	switch {
	case cl.Spec.KubernetesVersion != cl.Status.KubernetesVersion,
		diffCP != "",
		diffW != "",
		diffBastion != "":
		return true
	default:
		return false
	}
}

func (r *ClusterReconciler) delete(ctx context.Context, cl *undistrov1.Cluster) error {
	log := r.Log
	cl.Status.Ready = false
	if cl.Status.ClusterAPIRef == nil {
		cl.Status.Phase = undistrov1.DeletedPhase
		return nil
	}
	capiNM := types.NamespacedName{
		Name:      cl.Status.ClusterAPIRef.Name,
		Namespace: cl.Status.ClusterAPIRef.Namespace,
	}
	var capi clusterApi.Cluster
	err := r.Get(ctx, capiNM, &capi)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't get capi", "name", capiNM)
			return err
		}
		record.Event(cl, "ClusterDeleted", "Cluster deleted")
		cl.Status.Phase = undistrov1.DeletedPhase
		return nil
	}
	if capi.Status.Phase != string(clusterApi.ClusterPhaseDeleting) {
		if err := r.Delete(ctx, &capi); err != nil {
			if client.IgnoreNotFound(err) != nil {
				log.Error(err, "couldn't delete capi", "name", capiNM)
				return err
			}
			record.Event(cl, "ClusterDeleted", "Cluster deleted")
			cl.Status.Phase = undistrov1.DeletedPhase
			return nil
		}
	}
	cl.Status.Phase = undistrov1.DeletingPhase
	return nil
}

func (r *ClusterReconciler) init(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) error {
	opts := uclient.InitOptions{
		Kubeconfig: uclient.Kubeconfig{
			RestConfig: r.RestConfig,
		},
		InfrastructureProviders: []string{cl.Spec.InfrastructureProvider.NameVersion()},
		TargetNamespace:         "undistro-system",
		LogUsageInstructions:    false,
	}
	bp, cp := cl.GetManagedProvidersInfra()
	if bp != nil {
		opts.BootstrapProviders = bp
	}
	if cp != nil {
		opts.ControlPlaneProviders = cp
	}
	components, err := c.Init(opts)
	if err != nil {
		return err
	}
	if len(components) == 0 {
		var comp uclient.Components
		comp, err = c.GetProviderComponents(cl.Spec.InfrastructureProvider.Name, undistrov1.InfrastructureProviderType, uclient.ComponentsOptions{
			TargetNamespace: "undistro-system",
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
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
	cl.Status.Phase = undistrov1.InitializedPhase
	cl.Status.KubernetesVersion = cl.Spec.KubernetesVersion
	cl.Status.WorkerNodes = cl.Spec.WorkerNodes
	cl.Status.ControlPlaneNode = cl.Spec.ControlPlaneNode
	cl.Status.InfrastructureName = cl.Spec.InfrastructureProvider.Name
	cl.Status.TotalWorkerPools = int64(len(cl.Spec.WorkerNodes))
	cl.Status.BastionConfig = cl.Spec.Bastion
	cl.Status.TotalWorkerReplicas = 0
	for _, w := range cl.Spec.WorkerNodes {
		cl.Status.TotalWorkerReplicas += *w.Replicas
	}
	return nil
}

func (r *ClusterReconciler) config(ctx context.Context, cl *undistrov1.Cluster, opts template.Options, c uclient.Client) error {
	log := r.Log
	for _, ic := range cl.Status.InstalledComponents {
		comp, err := uclient.GetProvider(c, ic.Name, ic.Type)
		if err != nil {
			return err
		}
		preConfigFunc := comp.GetPreConfigFunc()
		if preConfigFunc != nil {
			log.Info("executing pre config func", "component", comp.Name())
			err = preConfigFunc(ctx, cl, c.GetVariables(), r.Client)
			if err != nil {
				return err
			}
		}
	}
	tpl := template.New(opts)
	buff := bytes.Buffer{}
	tplCluster := cl.DeepCopy()
	if tplCluster.Namespace == "" {
		tplCluster.Namespace = "default"
	}
	err := tpl.YAML(&buff, cl.Spec.InfrastructureProvider.Name, *tplCluster)
	if err != nil {
		return err
	}
	objs, err := yaml.ToUnstructured(buff.Bytes())
	if err != nil {
		return err
	}
	objs = utilresource.SortForCreate(objs)
	for _, o := range objs {
		isCluster := false
		if o.GetKind() == "Cluster" && o.GroupVersionKind().GroupVersion().String() == clusterApi.GroupVersion.String() {
			isCluster = true
			err = ctrl.SetControllerReference(cl, &o, r.Scheme)
			if err != nil {
				return errors.Errorf("couldn't set reference: %v", err)
			}
		}
		err = util.CreateOrUpdate(ctx, r.Client, o)
		if err != nil {
			return err
		}
		if isCluster {
			cl.Status.ClusterAPIRef = &corev1.ObjectReference{
				Kind:            o.GetKind(),
				Namespace:       o.GetNamespace(),
				Name:            o.GetName(),
				UID:             o.GetUID(),
				ResourceVersion: o.GetResourceVersion(),
				APIVersion:      o.GetAPIVersion(),
			}
		}
		r.EventWatcher <- record.WatcherResource{
			Cluster: cl,
			Obj:     &o,
		}
	}
	cl.Status.Phase = undistrov1.ProvisioningPhase
	return nil
}

func (r *ClusterReconciler) upgrade(ctx context.Context, cl *undistrov1.Cluster, uc uclient.Client, opts template.Options) (ctrl.Result, error) {
	if cl.Status.ClusterAPIRef == nil {
		return ctrl.Result{}, errors.New("cluster API reference is nil")
	}
	capi := clusterApi.Cluster{}
	nm := types.NamespacedName{
		Name:      cl.Status.ClusterAPIRef.Name,
		Namespace: cl.Status.ClusterAPIRef.Namespace,
	}
	if err := r.Get(ctx, nm, &capi); err != nil {
		return ctrl.Result{}, err
	}
	if cl.Namespace == "" {
		cl.Namespace = "default"
	}
	tpl := template.New(opts)
	buff := bytes.Buffer{}
	if err := tpl.YAML(&buff, cl.Spec.InfrastructureProvider.Name, *cl); err != nil {
		return ctrl.Result{}, err
	}
	objs, err := yaml.ToUnstructured(buff.Bytes())
	if err != nil {
		return ctrl.Result{}, err
	}
	objs = util.ReverseObjs(utilresource.SortForCreate(objs))
	for _, o := range objs {
		if o.GetKind() == "Cluster" && o.GroupVersionKind().GroupVersion().String() == clusterApi.GroupVersion.String() {
			// skip capi because is imutable, we will update provider cluster
			continue
		}
		err = util.CreateOrUpdate(ctx, r.Client, o)
		if err != nil {
			record.Warn(cl, "FailedUpgrade", "failed to upgrade Cluster API objects")
			return ctrl.Result{}, err
		}
	}
	cl.Status.KubernetesVersion = cl.Spec.KubernetesVersion
	cl.Status.WorkerNodes = cl.Spec.WorkerNodes
	cl.Status.ControlPlaneNode = cl.Spec.ControlPlaneNode
	cl.Status.InfrastructureName = cl.Spec.InfrastructureProvider.Name
	cl.Status.TotalWorkerPools = int64(len(cl.Spec.WorkerNodes))
	cl.Status.BastionConfig = cl.Spec.Bastion
	cl.Status.TotalWorkerReplicas = 0
	for _, w := range cl.Spec.WorkerNodes {
		cl.Status.TotalWorkerReplicas += *w.Replicas
	}
	record.Event(cl, "ClusterUpgrade", "Cluster API object has updated")
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) provisioning(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) (ctrl.Result, error) {
	log := r.Log
	log.Info("provisioning cluster", "name", cl.Name, "namespace", cl.Namespace)
	var clusterList clusterApi.ClusterList
	if err := r.List(ctx, &clusterList, client.InNamespace(cl.Namespace), client.MatchingFields{jobOwnerKey: cl.Name}); err != nil {
		return ctrl.Result{}, err
	}
	if len(clusterList.Items) != 1 {
		return ctrl.Result{}, errors.Errorf("has more than one Cluster API Cluster owned by %s/%s", cl.Namespace, cl.Name)
	}
	capi := &clusterList.Items[0]
	if undistrov1.ClusterPhase(capi.Status.Phase) == undistrov1.FailedPhase {
		cl.Status.Phase = undistrov1.FailedPhase
		return ctrl.Result{}, nil
	}
	if capi.Status.ControlPlaneInitialized && !capi.Status.ControlPlaneReady && cl.Spec.CniName != undistrov1.ProviderCNI {
		if err := r.installCNI(ctx, cl, c); err != nil {
			return ctrl.Result{}, err
		}
		record.Event(cl, "CNIInstalled", "CNI installed")
	}
	bastionIP, err := r.getClusterBastionIP(ctx, cl, capi)
	if err != nil {
		return ctrl.Result{}, err
	}
	if capi.Status.ControlPlaneReady && capi.Status.InfrastructureReady && undistrov1.ClusterPhase(capi.Status.Phase) == undistrov1.ProvisionedPhase {
		if cl.GetBastion().Enabled && bastionIP == "" {
			return ctrl.Result{Requeue: true}, nil
		}
		cl.Status.Phase = undistrov1.ProvisionedPhase
		cl.Status.Ready = true
		cl.Status.BastionPublicIP = bastionIP
		record.Event(cl, "ClusterReady", "Cluster ready")
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) getClusterBastionIP(ctx context.Context, cl *undistrov1.Cluster, capi *clusterApi.Cluster) (string, error) {
	var ref *corev1.ObjectReference
	if capi.Spec.InfrastructureRef == nil || capi.Spec.ControlPlaneRef == nil {
		return "", errors.New("nil provider cluster reference")
	}
	if cl.IsManaged() {
		ref = capi.Spec.ControlPlaneRef
	} else {
		ref = capi.Spec.InfrastructureRef
	}
	nm := client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
	o := unstructured.Unstructured{}
	o.SetGroupVersionKind(ref.GroupVersionKind())
	err := r.Get(ctx, nm, &o)
	if err != nil {
		return "", client.IgnoreNotFound(err)
	}
	ip, _, err := unstructured.NestedString(o.Object, "status", "bastion", "publicIp")
	return ip, err
}

func (r *ClusterReconciler) installCNI(ctx context.Context, cl *undistrov1.Cluster, c uclient.Client) error {
	log := r.Log
	cniAddr := cl.GetCNITemplateURL()
	if cniAddr == "" {
		return errors.Errorf("CNI %s is not supported", cl.Spec.CniName)
	}
	log.Info("getting CNI", "name", cl.Spec.CniName, "URL", cniAddr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cniAddr, nil)
	if err != nil {
		return err
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	objs, err := yaml.ToUnstructured(byt)
	if err != nil {
		return err
	}
	wc, err := c.GetWorkloadCluster(uclient.Kubeconfig{
		RestConfig: r.RestConfig,
	})
	if err != nil {
		return err
	}
	workloadCfg, err := wc.GetRestConfig(cl.Name, cl.Namespace)
	if err != nil {
		return err
	}
	workloadClient, err := client.New(workloadCfg, client.Options{Scheme: r.Scheme})
	if err != nil {
		return err
	}
	objs = utilresource.SortForCreate(objs)
	for _, o := range objs {
		err = util.CreateOrUpdate(ctx, workloadClient, o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, opts controller.Options) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &clusterApi.Cluster{}, jobOwnerKey, func(rawObj client.Object) []string {
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
		Watches(
			&source.Kind{Type: &clusterApi.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.capiToUndistro),
		).
		Complete(r)
}

func (r *ClusterReconciler) capiToUndistro(o client.Object) []ctrl.Request {
	ctx := context.TODO()
	c, ok := o.(*clusterApi.Cluster)
	if !ok {
		r.Log.Error(nil, fmt.Sprintf("expected a Cluster but got a %T", o))
		return nil
	}
	if c.Status.Phase == "" {
		return nil
	}
	nm := types.NamespacedName{
		Name:      c.Name,
		Namespace: c.Namespace,
	}
	uc := undistrov1.Cluster{}
	if err := r.Get(ctx, nm, &uc); err != nil {
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "couldn't get undistro cluster")
		}
		return nil
	}
	if uc.Status.Phase == undistrov1.ClusterPhase(c.Status.Phase) {
		return nil
	}
	uc.Status.Phase = undistrov1.ClusterPhase(c.Status.Phase)
	if err := r.Status().Update(ctx, &uc); client.IgnoreNotFound(err) != nil {
		r.Log.Error(err, "couldn't update status", "name", uc.Name, "namespace", uc.Namespace)
		return nil
	}
	return []ctrl.Request{
		{
			NamespacedName: nm,
		},
	}
}
