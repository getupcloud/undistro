/*
Copyright 2020 The UnDistro authors

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

package controllers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/template"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/uuid"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capicp "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	capiexp "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cl := appv1alpha1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, &cl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log := r.Log.WithValues("cluster", req.NamespacedName)
	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(&cl, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		patchOpts := []patch.Option{}
		if err == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		patchErr := patchHelper.Patch(ctx, &cl, patchOpts...)
		if patchErr != nil {
			err = kerrors.NewAggregate([]error{patchErr, err})
		}
	}()
	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&cl, meta.Finalizer) {
		controllerutil.AddFinalizer(&cl, meta.Finalizer)
		return ctrl.Result{}, nil
	}
	if cl.Spec.Paused {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	capiCluster := capi.Cluster{}
	err = r.Get(ctx, client.ObjectKeyFromObject(&cl), &capiCluster)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	}
	if !cl.DeletionTimestamp.IsZero() {
		log.Info("Deleting")
		return r.reconcileDelete(ctx, log, cl)
	}
	cl, result, err := r.reconcile(ctx, log, cl, capiCluster)
	return result, err
}

func (r *ClusterReconciler) templateVariables(ctx context.Context, capiCluster *capi.Cluster, cl *appv1alpha1.Cluster) (map[string]interface{}, error) {
	vars := make(map[string]interface{})
	v := make(map[string]interface{})
	err := template.SetVariablesFromEnvVar(ctx, template.VariablesInput{
		ClientSet:      r.Client,
		NamespacedName: client.ObjectKeyFromObject(cl),
		Variables:      v,
		EnvVars:        cl.Spec.InfrastructureProvider.Env,
	})
	if err != nil {
		return nil, err
	}
	vars["Cluster"] = cl
	vars["ENV"] = v
	validDiff := true
	labels := cl.GetLabels()
	_, moved := labels[meta.LabelUndistroMoved]
	if moved && !cl.Spec.InfrastructureProvider.IsManaged() && cl.Status.LastUsedUID == "" {
		validDiff = false
		cp := capicp.KubeadmControlPlane{}
		err := r.Get(ctx, client.ObjectKeyFromObject(cl), &cp)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if !apierrors.IsNotFound(err) {
			split := strings.Split(cp.Spec.InfrastructureTemplate.Name, "-")
			cl.Status.LastUsedUID = split[len(split)-1]
		}
	}
	if r.hasDiff(ctx, cl) && validDiff {
		cl.Status.LastUsedUID = string(uuid.NewUUID())
	}
	return vars, nil
}

func (r *ClusterReconciler) getBastionIP(ctx context.Context, log logr.Logger, cl appv1alpha1.Cluster, capiCluster capi.Cluster) (string, error) {
	ref := capiCluster.Spec.InfrastructureRef
	if cl.Spec.InfrastructureProvider.IsManaged() {
		ref = capiCluster.Spec.ControlPlaneRef
	}
	if ref != nil {
		key := client.ObjectKey{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		}
		o := unstructured.Unstructured{}
		o.SetGroupVersionKind(ref.GroupVersionKind())
		err := r.Get(ctx, key, &o)
		if err != nil {
			return "", err
		}
		ip, _, err := unstructured.NestedString(o.Object, "status", "bastion", "publicIp")
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return "", nil
}

func (r *ClusterReconciler) reconcile(ctx context.Context, log logr.Logger, cl appv1alpha1.Cluster, capiCluster capi.Cluster) (appv1alpha1.Cluster, ctrl.Result, error) {
	cl.Status.TotalWorkerPools = int32(len(cl.Spec.Workers))
	cl.Status.TotalWorkerReplicas = 0
	for _, w := range cl.Spec.Workers {
		cl.Status.TotalWorkerReplicas += *w.Replicas
	}
	for _, cond := range cl.Status.Conditions {
		meta.SetResourceCondition(&cl, cond.Type, cond.Status, cond.Reason, cond.Message)
	}
	if capiCluster.Status.ControlPlaneInitialized && !capiCluster.Status.ControlPlaneReady && !cl.Spec.InfrastructureProvider.IsManaged() {
		log.Info("installing calico")
		err := r.installCNI(ctx, cl)
		if err != nil {
			meta.SetResourceCondition(&cl, meta.CNIInstalledCondition, metav1.ConditionFalse, meta.CNIInstalledFailedReason, err.Error())
			return cl, ctrl.Result{}, err
		}
		meta.SetResourceCondition(&cl, meta.CNIInstalledCondition, metav1.ConditionTrue, meta.CNIInstalledSuccessReason, "calico installed")
	}
	if cl.Spec.Bastion != nil {
		if *cl.Spec.Bastion.Enabled && cl.Status.BastionPublicIP == "" {
			var err error
			cl.Status.BastionPublicIP, err = r.getBastionIP(ctx, log, cl, capiCluster)
			if err != nil {
				return appv1alpha1.ClusterNotReady(cl, meta.WaitProvisionReason, err.Error()), ctrl.Result{Requeue: true}, nil
			}
		}
	}
	if capiCluster.Spec.ClusterNetwork != nil {
		if !reflect.DeepEqual(*capiCluster.Spec.ClusterNetwork, capi.ClusterNetwork{}) && reflect.DeepEqual(cl.Spec.Network.ClusterNetwork, capi.ClusterNetwork{}) {
			cl.Spec.Network.ClusterNetwork = *capiCluster.Spec.ClusterNetwork
		}
	}
	if !reflect.DeepEqual(capiCluster.Spec.ControlPlaneEndpoint, capi.APIEndpoint{}) && reflect.DeepEqual(cl.Spec.ControlPlane.Endpoint, capi.APIEndpoint{}) {
		cl.Spec.ControlPlane.Endpoint = capiCluster.Spec.ControlPlaneEndpoint
	}
	vars, err := r.templateVariables(ctx, &capiCluster, &cl)
	if err != nil {
		return appv1alpha1.ClusterNotReady(cl, meta.TemplateAppliedFailed, err.Error()), ctrl.Result{}, err
	}
	tpl := template.New(template.Options{
		Directory: "clustertemplates",
	})
	buff := &bytes.Buffer{}
	err = tpl.YAML(buff, cl.GetTemplate(), vars)
	if err != nil {
		return appv1alpha1.ClusterNotReady(cl, meta.TemplateAppliedFailed, err.Error()), ctrl.Result{}, err
	}
	objs, err := util.ToUnstructured(buff.Bytes())
	if err != nil {
		return appv1alpha1.ClusterNotReady(cl, meta.TemplateAppliedFailed, err.Error()), ctrl.Result{}, err
	}
	for _, o := range objs {
		if o.GetAPIVersion() == capi.GroupVersion.String() && o.GetKind() == "Cluster" {
			err = ctrl.SetControllerReference(&cl, &o, scheme.Scheme)
			if err != nil {
				return cl, ctrl.Result{}, err
			}
		}
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			_, err = util.CreateOrUpdate(ctx, r.Client, &o)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return cl, ctrl.Result{}, err
		}
	}
	cl.Status.KubernetesVersion = cl.Spec.KubernetesVersion
	cl.Status.ControlPlane = *cl.Spec.ControlPlane
	cl.Status.Workers = cl.Spec.Workers
	cl.Status.BastionConfig = cl.Spec.Bastion
	if capiCluster.Status.InfrastructureReady && capiCluster.Status.ControlPlaneReady && capiCluster.Status.GetTypedPhase() == capi.ClusterPhaseProvisioned {
		err = r.reconcileNodes(ctx, cl, capiCluster)
		if err != nil {
			return appv1alpha1.ClusterNotReady(cl, meta.ReconcileNodesFailed, err.Error()), ctrl.Result{}, err
		}
		cl = appv1alpha1.ClusterReady(cl)
		return cl, ctrl.Result{}, nil
	}
	return appv1alpha1.ClusterNotReady(cl, meta.WaitProvisionReason, "wait cluster to be provisioned"), ctrl.Result{Requeue: true}, nil
}

func (r *ClusterReconciler) reconcileNodes(ctx context.Context, cl appv1alpha1.Cluster, capiCluster capi.Cluster) error {
	wc, err := kube.NewClusterClient(ctx, r.Client, cl.Name, cl.GetNamespace())
	if err != nil {
		return err
	}
	if !cl.Spec.InfrastructureProvider.IsManaged() {
		if len(cl.Spec.ControlPlane.Labels) > 0 || len(cl.Spec.ControlPlane.Taints) > 0 {
			cp, _, err := util.GetMachinesForCluster(ctx, r.Client, &capiCluster)
			if err != nil {
				return err
			}
			for _, m := range cp.Items {
				key := client.ObjectKey{
					Name: m.Status.NodeRef.Name,
				}
				node := corev1.Node{}
				err = wc.Get(ctx, key, &node)
				if err != nil {
					return err
				}
				if node.Labels == nil {
					node.Labels = make(map[string]string)
				}
				for k, v := range cl.Spec.ControlPlane.Labels {
					node.Labels[k] = v
				}
				node.Spec.Taints = append(node.Spec.Taints, cl.Spec.ControlPlane.Taints...)
				node.Spec.Taints = util.RemoveDuplicateTaints(node.Spec.Taints)
				node.TypeMeta = metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Node",
				}
				_, err = util.CreateOrUpdate(ctx, wc, &node)
				if err != nil {
					return err
				}
			}
		}
	}
	// workers
	mpList := capiexp.MachinePoolList{}
	err = r.List(ctx, &mpList, client.HasLabels{capi.ClusterLabelName})
	if err != nil {
		return err
	}
	for _, mp := range mpList.Items {
		w, err := cl.GetWorkerRefByMachinePool(mp.Name)
		if err != nil {
			return err
		}
		if len(w.Labels) == 0 && len(w.Taints) == 0 {
			continue
		}
		for _, ref := range mp.Status.NodeRefs {
			key := client.ObjectKey{
				Name: ref.Name,
			}
			node := corev1.Node{}
			err = wc.Get(ctx, key, &node)
			if err != nil {
				return err
			}
			if node.Labels == nil {
				node.Labels = make(map[string]string)
			}
			for k, v := range w.Labels {
				node.Labels[k] = v
			}
			node.Spec.Taints = append(node.Spec.Taints, w.Taints...)
			node.Spec.Taints = util.RemoveDuplicateTaints(node.Spec.Taints)
			node.TypeMeta = metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Node",
			}
			_, err = util.CreateOrUpdate(ctx, wc, &node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ClusterReconciler) hasDiff(ctx context.Context, cl *appv1alpha1.Cluster) bool {
	if cl.Spec.KubernetesVersion != cl.Status.KubernetesVersion {
		return true
	}
	if !cl.Spec.InfrastructureProvider.IsManaged() && cl.Spec.ControlPlane != nil {
		if *cl.Spec.ControlPlane.Replicas != *cl.Status.ControlPlane.Replicas {
			return true
		}
		if cl.Spec.ControlPlane.MachineType != cl.Status.ControlPlane.MachineType {
			return true
		}
	}
	return false
}

func (r *ClusterReconciler) installCNI(ctx context.Context, cl appv1alpha1.Cluster) error {
	const addr = "https://docs.projectcalico.org/manifests/calico.yaml"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
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
	buff := bytes.Buffer{}
	_, err = io.Copy(&buff, resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unable to get calico: %s", buff.String())
	}
	objs, err := util.ToUnstructured(buff.Bytes())
	if err != nil {
		return err
	}
	workloadClient, err := kube.NewClusterClient(ctx, r.Client, cl.Name, cl.GetNamespace())
	if err != nil {
		return err
	}
	for _, o := range objs {
		_, err = util.CreateOrUpdate(ctx, workloadClient, &o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, logger logr.Logger, cl appv1alpha1.Cluster) (ctrl.Result, error) {
	capiCluster := capi.Cluster{}
	err := r.Get(ctx, client.ObjectKeyFromObject(&cl), &capiCluster)
	if apierrors.IsNotFound(err) {
		controllerutil.RemoveFinalizer(&cl, meta.Finalizer)
		_, err = util.CreateOrUpdate(ctx, r.Client, &cl)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{Requeue: true}, r.Delete(ctx, &capiCluster)
}

func (r *ClusterReconciler) capiToUndistro(o client.Object) []ctrl.Request {
	capiCluster, ok := o.(*capi.Cluster)
	if !ok {
		return nil
	}
	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKeyFromObject(capiCluster),
		},
	}
}

func (r *ClusterReconciler) mpToUndistro(o client.Object) []ctrl.Request {
	capiMP, ok := o.(*capiexp.MachinePool)
	if !ok {
		return nil
	}
	if capiMP.Labels == nil {
		return nil
	}
	name, ok := capiMP.Labels[capi.ClusterLabelName]
	if !ok {
		return nil
	}
	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      name,
				Namespace: capiMP.GetNamespace(),
			},
		},
	}
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Cluster{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Watches(
			&source.Kind{
				Type: &capi.Cluster{},
			},
			handler.EnqueueRequestsFromMapFunc(r.capiToUndistro),
		).
		Watches(
			&source.Kind{
				Type: &capiexp.MachinePool{},
			},
			handler.EnqueueRequestsFromMapFunc(r.mpToUndistro),
		).
		Complete(r)
}
