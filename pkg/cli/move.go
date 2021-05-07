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
package cli

import (
	"context"
	"fmt"
	"os"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/graph"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MoveOptions struct {
	ConfigPath  string
	ClusterName string
	Namespace   string
	genericclioptions.IOStreams
}

func NewMoveOptions(streams genericclioptions.IOStreams) *MoveOptions {
	return &MoveOptions{
		IOStreams: streams,
	}
}

func (o *MoveOptions) Complete(f *ConfigFlags, cmd *cobra.Command, args []string) error {
	o.ConfigPath = *f.ConfigFile
	var err error
	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		// do nothing
	case 1:
		o.ClusterName = args[0]
	default:
		return cmdutil.UsageErrorf(cmd, "%s", "too many arguments")
	}
	return nil
}

func (o *MoveOptions) Validate() error {
	if o.ConfigPath != "" {
		_, err := os.Stat(o.ConfigPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *MoveOptions) RunMove(f cmdutil.Factory, cmd *cobra.Command) error {
	log := log.Log
	cfg := Config{}
	if o.ConfigPath != "" {
		err := viper.Unmarshal(&cfg)
		if err != nil {
			return errors.Errorf("unable to unmarshal config: %v", err)
		}
	}
	key := client.ObjectKey{
		Namespace: o.Namespace,
		Name:      o.ClusterName,
	}
	iopts := &InstallOptions{
		ConfigPath:  o.ConfigPath,
		ClusterName: key.String(),
		IOStreams:   o.IOStreams,
	}
	err := iopts.RunInstall(f, cmd)
	if err != nil {
		return err
	}
	localCfg, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	localClient, err := client.New(localCfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}
	byt, err := kubeconfig.FromSecret(cmd.Context(), localClient, key)
	if err != nil {
		return err
	}
	restGetter := kube.NewMemoryRESTClientGetter(byt, o.Namespace)
	remoteCfg, err := restGetter.ToRESTConfig()
	if err != nil {
		return err
	}
	remoteClient, err := client.New(remoteCfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}
	objectGraph := graph.NewObjectGraph(localClient, o.IOStreams)
	// Gets all the types defines by the CRDs installed by undistro plus the ConfigMap/Secret core types.
	err = objectGraph.GetDiscoveryTypes()
	if err != nil {
		return err
	}
	if err := objectGraph.Discovery(o.Namespace); err != nil {
		return err
	}
	// Check whether nodes are not included in GVK considered for move
	objectGraph.CheckVirtualNode()
	clusters := objectGraph.GetClusters()
	fmt.Fprintf(o.Out, "Moving %d clusters\n", len(clusters))
	errList := []error{}
	for i := range clusters {
		cluster := clusters[i]
		clusterObj := &appv1alpha1.Cluster{}
		if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return getClusterObj(cmd.Context(), localClient, cluster, clusterObj)
		}); err != nil {
			return err
		}

		if !meta.InReadyCondition(clusterObj.Status.Conditions) {
			errList = append(errList, errors.Errorf("cannot start the move operation while %q %s/%s is still provisioning the cluster", clusterObj.GroupVersionKind(), clusterObj.GetNamespace(), clusterObj.GetName()))
		}
	}
	if len(errList) > 0 {
		return kerrors.NewAggregate(errList)
	}

	machines := objectGraph.GetMachines()
	for i := range machines {
		machine := machines[i]
		machineObj := &clusterv1.Machine{}
		if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return getMachineObj(cmd.Context(), localClient, machine, machineObj)
		}); err != nil {
			return err
		}

		if machineObj.Status.NodeRef == nil {
			errList = append(errList, errors.Errorf("cannot start the move operation while %q %s/%s is still provisioning the node", machineObj.GroupVersionKind(), machineObj.GetNamespace(), machineObj.GetName()))
		}
	}
	if len(errList) > 0 {
		return kerrors.NewAggregate(errList)
	}
	// Sets the pause field on the Cluster object in the source management cluster, so the controllers stop reconciling it.
	log.V(1).Info("Pausing the source cluster")
	if err := setClusterPause(cmd.Context(), localClient, clusters, true); err != nil {
		return err
	}
	// Ensure all the expected target namespaces are in place before creating objects.
	log.V(1).Info("Creating target namespaces, if missing")
	if err := o.ensureNamespaces(cmd.Context(), objectGraph, remoteClient); err != nil {
		return err
	}
	// Define the move sequence by processing the ownerReference chain, so we ensure that a Kubernetes object is moved only after its owners.
	// The sequence is bases on object graph nodes, each one representing a Kubernetes object; nodes are grouped, so bulk of nodes can be moved in parallel. e.g.
	// - All the Clusters should be moved first (group 1, processed in parallel)
	// - All the MachineDeployments should be moved second (group 1, processed in parallel)
	// - then all the MachineSets, then all the Machines, etc.
	moveSequence := graph.GetMoveSequence(objectGraph)
	fmt.Fprintln(o.Out, "Creating objects in target clusters")
	for groupIndex := 0; groupIndex < len(moveSequence.Groups); groupIndex++ {
		if err := o.createGroup(cmd.Context(), moveSequence.GetGroup(groupIndex), remoteClient, localClient); err != nil {
			return err
		}
	}

	// Delete all objects group by group in reverse order.
	fmt.Fprintln(o.Out, "Deleting objects from the source cluster")
	for groupIndex := len(moveSequence.Groups) - 1; groupIndex >= 0; groupIndex-- {
		if err := o.deleteGroup(cmd.Context(), moveSequence.GetGroup(groupIndex), localClient); err != nil {
			return err
		}
	}

	// Reset the pause field on the Cluster object in the target management cluster, so the controllers start reconciling it.
	log.V(1).Info("Resuming the target cluster")
	if err := setClusterPause(cmd.Context(), remoteClient, clusters, false); err != nil {
		return err
	}
	return nil
}

func getClusterObj(ctx context.Context, proxy client.Client, cluster *graph.Node, clusterObj *appv1alpha1.Cluster) error {
	clusterObjKey := client.ObjectKey{
		Namespace: cluster.Identity.Namespace,
		Name:      cluster.Identity.Name,
	}

	if err := proxy.Get(ctx, clusterObjKey, clusterObj); err != nil {
		return errors.Wrapf(err, "error reading %q %s/%s",
			clusterObj.GroupVersionKind(), clusterObj.GetNamespace(), clusterObj.GetName())
	}
	return nil
}

// getMachineObj retrieves the the machineObj corresponding to a node with type Machine.
func getMachineObj(ctx context.Context, proxy client.Client, machine *graph.Node, machineObj *clusterv1.Machine) error {
	machineObjKey := client.ObjectKey{
		Namespace: machine.Identity.Namespace,
		Name:      machine.Identity.Name,
	}

	if err := proxy.Get(ctx, machineObjKey, machineObj); err != nil {
		return errors.Wrapf(err, "error reading %q %s/%s",
			machineObj.GroupVersionKind(), machineObj.GetNamespace(), machineObj.GetName())
	}
	return nil
}

// setClusterPause sets the paused field on nodes referring to Cluster objects.
func setClusterPause(ctx context.Context, proxy client.Client, clusters []*graph.Node, value bool) error {
	log := log.Log
	patch := client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf("{\"spec\":{\"paused\":%t}}", value)))

	for i := range clusters {
		cluster := clusters[i]
		log.V(5).Info("Set Cluster.Spec.Paused", "Paused", value, "Cluster", cluster.Identity.Name, "Namespace", cluster.Identity.Namespace)

		// Nb. The operation is wrapped in a retry loop to make setClusterPause more resilient to unexpected conditions.
		if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return patchCluster(ctx, proxy, cluster, patch)
		}); err != nil {
			return err
		}
	}
	return nil
}

// ensureNamespaces ensures all the expected target namespaces are in place before creating objects.
func (o *MoveOptions) ensureNamespaces(ctx context.Context, graph *graph.ObjectGraph, toProxy client.Client) error {
	namespaces := sets.NewString()
	for _, node := range graph.GetMoveNodes() {

		// ignore global/cluster-wide objects
		if node.IsGlobal {
			continue
		}

		namespace := node.Identity.Namespace

		// If the namespace was already processed, skip it.
		if namespaces.Has(namespace) {
			continue
		}
		namespaces.Insert(namespace)

		if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return o.ensureNamespace(ctx, toProxy, namespace)
		}); err != nil {
			return err
		}
	}

	return nil
}

// ensureNamespace ensures a target namespaces is in place before creating objects.
func (o *MoveOptions) ensureNamespace(ctx context.Context, toProxy client.Client, namespace string) error {
	log := log.Log
	// Otherwise check if namespace exists (also dealing with RBAC restrictions).
	ns := &corev1.Namespace{}
	key := client.ObjectKey{
		Name: namespace,
	}

	err := toProxy.Get(ctx, key, ns)
	if err == nil {
		return nil
	}
	if apierrors.IsForbidden(err) {
		namespaces := &corev1.NamespaceList{}
		namespaceExists := false
		for {
			if err := toProxy.List(ctx, namespaces, client.Continue(namespaces.Continue)); err != nil {
				return err
			}

			for _, ns := range namespaces.Items {
				if ns.Name == namespace {
					namespaceExists = true
					break
				}
			}

			if namespaces.Continue == "" {
				break
			}
		}
		if namespaceExists {
			return nil
		}
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	// If the namespace does not exists, create it.
	ns = &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	log.V(1).Info("Creating", ns.Kind, ns.Name)
	if err := toProxy.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// createGroup creates all the Kubernetes objects into the target management cluster corresponding to the object graph nodes in a moveGroup.
func (o *MoveOptions) createGroup(ctx context.Context, group graph.MoveGroup, toProxy client.Client, fromProxy client.Client) error {
	errList := []error{}
	for i := range group {
		nodeToCreate := group[i]

		// Creates the Kubernetes object corresponding to the nodeToCreate.
		// Nb. The operation is wrapped in a retry loop to make move more resilient to unexpected conditions.
		err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return o.createTargetObject(ctx, nodeToCreate, toProxy, fromProxy)
		})
		if err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return kerrors.NewAggregate(errList)
	}

	return nil
}

// deleteGroup deletes all the Kubernetes objects from the source management cluster corresponding to the object graph nodes in a moveGroup.
func (o *MoveOptions) deleteGroup(ctx context.Context, group graph.MoveGroup, fromProxy client.Client) error {
	errList := []error{}
	for i := range group {
		nodeToDelete := group[i]

		// Don't delete cluster-wide nodes
		if nodeToDelete.IsGlobal {
			continue
		}

		// Delete the Kubernetes object corresponding to the current node.
		// Nb. The operation is wrapped in a retry loop to make move more resilient to unexpected conditions.
		err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return o.deleteSourceObject(ctx, nodeToDelete, fromProxy)
		})

		if err != nil {
			errList = append(errList, err)
		}
	}

	return kerrors.NewAggregate(errList)
}

var (
	removeFinalizersPatch = client.RawPatch(types.MergePatchType, []byte("{\"metadata\":{\"finalizers\":[]}}"))
)

// deleteSourceObject deletes the Kubernetes object corresponding to the node from the source management cluster, taking care of removing all the finalizers so
// the objects gets immediately deleted (force delete).
func (o *MoveOptions) deleteSourceObject(ctx context.Context, nodeToDelete *graph.Node, fromProxy client.Client) error {
	log := log.Log
	log.V(1).Info("Deleting", nodeToDelete.Identity.Kind, nodeToDelete.Identity.Name, "Namespace", nodeToDelete.Identity.Namespace)

	// Get the source object
	sourceObj := &unstructured.Unstructured{}
	sourceObj.SetAPIVersion(nodeToDelete.Identity.APIVersion)
	sourceObj.SetKind(nodeToDelete.Identity.Kind)
	sourceObjKey := client.ObjectKey{
		Namespace: nodeToDelete.Identity.Namespace,
		Name:      nodeToDelete.Identity.Name,
	}

	if err := fromProxy.Get(ctx, sourceObjKey, sourceObj); err != nil {
		if apierrors.IsNotFound(err) {
			//If the object is already deleted, move on.
			log.V(5).Info("Object already deleted, skipping delete for", nodeToDelete.Identity.Kind, nodeToDelete.Identity.Name, "Namespace", nodeToDelete.Identity.Namespace)
			return nil
		}
		return errors.Wrapf(err, "error reading %q %s/%s",
			sourceObj.GroupVersionKind(), sourceObj.GetNamespace(), sourceObj.GetName())
	}

	if len(sourceObj.GetFinalizers()) > 0 {
		if err := fromProxy.Patch(ctx, sourceObj, removeFinalizersPatch); err != nil {
			return errors.Wrapf(err, "error removing finalizers from %q %s/%s",
				sourceObj.GroupVersionKind(), sourceObj.GetNamespace(), sourceObj.GetName())
		}
	}

	if err := fromProxy.Delete(ctx, sourceObj); err != nil {
		return errors.Wrapf(err, "error deleting %q %s/%s",
			sourceObj.GroupVersionKind(), sourceObj.GetNamespace(), sourceObj.GetName())
	}

	return nil
}

// createTargetObject creates the Kubernetes object in the target Management cluster corresponding to the object graph node, taking care of restoring the OwnerReference with the owner nodes, if any.
func (o *MoveOptions) createTargetObject(ctx context.Context, nodeToCreate *graph.Node, toProxy client.Client, fromProxy client.Client) error {
	log := log.Log
	log.V(1).Info("Creating", nodeToCreate.Identity.Kind, nodeToCreate.Identity.Name, "Namespace", nodeToCreate.Identity.Namespace)

	// Get the source object
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(nodeToCreate.Identity.APIVersion)
	obj.SetKind(nodeToCreate.Identity.Kind)
	objKey := client.ObjectKey{
		Namespace: nodeToCreate.Identity.Namespace,
		Name:      nodeToCreate.Identity.Name,
	}

	if err := fromProxy.Get(ctx, objKey, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return err
		}
		return errors.Wrapf(err, "error reading %q %s/%s",
			obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	// New objects cannot have a specified resource version. Clear it out.
	obj.SetResourceVersion("")

	// Removes current OwnerReferences
	obj.SetOwnerReferences(nil)

	// Recreate all the OwnerReferences using the newUID of the owner nodes.
	if len(nodeToCreate.Owners) > 0 {
		ownerRefs := []metav1.OwnerReference{}
		for ownerNode := range nodeToCreate.Owners {
			ownerRef := metav1.OwnerReference{
				APIVersion: ownerNode.Identity.APIVersion,
				Kind:       ownerNode.Identity.Kind,
				Name:       ownerNode.Identity.Name,
				UID:        ownerNode.NewUID, // Use the owner's newUID read from the target management cluster (instead of the UID read during discovery).
			}

			// Restores the attributes of the OwnerReference.
			if attributes, ok := nodeToCreate.Owners[ownerNode]; ok {
				ownerRef.Controller = attributes.Controller
				ownerRef.BlockOwnerDeletion = attributes.BlockOwnerDeletion
			}

			ownerRefs = append(ownerRefs, ownerRef)
		}
		obj.SetOwnerReferences(ownerRefs)

	}
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[meta.LabelUndistroMoved] = "true"
	obj.SetLabels(labels)
	if err := toProxy.Create(ctx, obj); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "error creating %q %s/%s",
				obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
		}

		// If the object already exists, try to update it.
		// Nb. This should not happen, but it is supported to make move more resilient to unexpected interrupt/restarts of the move process.
		log.V(5).Info("Object already exists, updating", nodeToCreate.Identity.Kind, nodeToCreate.Identity.Name, "Namespace", nodeToCreate.Identity.Namespace)

		// Retrieve the UID and the resource version for the update.
		existingTargetObj := &unstructured.Unstructured{}
		existingTargetObj.SetAPIVersion(obj.GetAPIVersion())
		existingTargetObj.SetKind(obj.GetKind())
		if err := toProxy.Get(ctx, objKey, existingTargetObj); err != nil {
			return errors.Wrapf(err, "error reading resource for %q %s/%s",
				existingTargetObj.GroupVersionKind(), existingTargetObj.GetNamespace(), existingTargetObj.GetName())
		}

		obj.SetUID(existingTargetObj.GetUID())
		obj.SetResourceVersion(existingTargetObj.GetResourceVersion())
		labels := obj.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[meta.LabelUndistroMoved] = "true"
		obj.SetLabels(labels)
		if err := toProxy.Update(ctx, obj); err != nil {
			return errors.Wrapf(err, "error updating %q %s/%s",
				obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
		}
	}

	// Stores the newUID assigned to the newly created object.
	nodeToCreate.NewUID = obj.GetUID()

	return nil
}

// patchCluster applies a patch to a node referring to a Cluster object.
func patchCluster(ctx context.Context, proxy client.Client, cluster *graph.Node, patch client.Patch) error {
	clusterObj := &appv1alpha1.Cluster{}
	clusterObjKey := client.ObjectKey{
		Namespace: cluster.Identity.Namespace,
		Name:      cluster.Identity.Name,
	}

	if err := proxy.Get(ctx, clusterObjKey, clusterObj); err != nil {
		return errors.Wrapf(err, "error reading %q %s/%s",
			clusterObj.GroupVersionKind(), clusterObj.GetNamespace(), clusterObj.GetName())
	}

	if err := proxy.Patch(ctx, clusterObj, patch); err != nil {
		return errors.Wrapf(err, "error pausing reconciliation for %q %s/%s",
			clusterObj.GroupVersionKind(), clusterObj.GetNamespace(), clusterObj.GetName())
	}

	return nil
}

func NewCmdMove(f *ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewMoveOptions(streams)
	cmd := &cobra.Command{
		Use:                   "move [cluster name]",
		DisableFlagsInUseLine: true,
		Short:                 "Move UnDistro resources to another cluster",
		Long: LongDesc(`Install UnDistro.
		IMove UnDistro resources to cluster passed as argument`),
		Example: Examples(`
		# Move
		undistro --config undistro-config.yaml move cool-product-cluster -n undistro-production
		`),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunMove(cmdutil.NewFactory(f), cmd))
		},
	}
	return cmd
}
