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
package graph

import (
	"context"
	"fmt"
	"strings"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Empty struct{}

type ownerReferenceAttributes struct {
	Controller         *bool
	BlockOwnerDeletion *bool
}

// Node defines a Node in the Kubernetes object graph that is visited during the discovery phase for the move operation.
type Node struct {
	Identity corev1.ObjectReference

	// Owners contains the list of nodes that are owned by the current node.
	Owners map[*Node]ownerReferenceAttributes

	// SoftOwners contains the list of nodes that are soft-owned by the current node.
	// E.g. secrets are soft-owned by a cluster via a naming convention, but without an explicit OwnerReference.
	SoftOwners map[*Node]Empty

	// ForceMove is set to true if the CRD of this object has the "move" label attached.
	// This ensures the node is moved, regardless of its owner refs.
	ForceMove bool

	// IsGlobal gets set to true if this object is a global resource (no namespace).
	IsGlobal bool

	// Virtual records if this node was discovered indirectly, e.g. by processing an OwnerRef, but not yet observed as a concrete object.
	Virtual bool

	//newID stores the new UID the objects gets once created in the target cluster.
	NewUID types.UID

	// TenantClusters define the list of Clusters which are tenant for the node, no matter if the node has a direct OwnerReference to the Cluster or if
	// the node is linked to a Cluster indirectly in the OwnerReference chain.
	TenantClusters map[*Node]Empty

	// TenantCRSs define the list of ClusterResourceSet which are tenant for the node, no matter if the node has a direct OwnerReference to the ClusterResourceSet or if
	// the node is linked to a ClusterResourceSet indirectly in the OwnerReference chain.
	TenantCRSs map[*Node]Empty
}

type discoveryTypeInfo struct {
	typeMeta  metav1.TypeMeta
	forceMove bool
}

// markObserved marks the fact that a node was observed as a concrete object.
func (n *Node) markObserved() {
	n.Virtual = false
}

func (n *Node) addOwner(owner *Node, attributes ownerReferenceAttributes) {
	n.Owners[owner] = attributes
}

func (n *Node) addSoftOwner(owner *Node) {
	n.SoftOwners[owner] = struct{}{}
}

func (n *Node) isOwnedBy(other *Node) bool {
	_, ok := n.Owners[other]
	return ok
}

func (n *Node) isSoftOwnedBy(other *Node) bool {
	_, ok := n.SoftOwners[other]
	return ok
}

// ObjectGraph manages the Kubernetes object graph that is generated during the discovery phase for the move operation.
type ObjectGraph struct {
	proxy     client.Client
	uidToNode map[types.UID]*Node
	types     map[string]*discoveryTypeInfo
	genericclioptions.IOStreams
}

func NewObjectGraph(proxy client.Client, streams genericclioptions.IOStreams) *ObjectGraph {
	return &ObjectGraph{
		proxy:     proxy,
		uidToNode: map[types.UID]*Node{},
		types:     map[string]*discoveryTypeInfo{},
		IOStreams: streams,
	}
}

// addObj adds a Kubernetes object to the object graph that is generated during the move discovery phase.
// During add, OwnerReferences are processed in order to create the dependency graph.
func (o *ObjectGraph) addObj(obj *unstructured.Unstructured) {
	// Adds the node to the Graph.
	newNode := o.objToNode(obj)

	// Process OwnerReferences; if the owner object doe not exists yet, create a virtual node as a placeholder for it.
	for _, ownerReference := range obj.GetOwnerReferences() {
		ownerNode, ok := o.uidToNode[ownerReference.UID]
		if !ok {
			ownerNode = o.ownerToVirtualNode(ownerReference, newNode.Identity.Namespace)
		}

		newNode.addOwner(ownerNode, ownerReferenceAttributes{
			Controller:         ownerReference.Controller,
			BlockOwnerDeletion: ownerReference.BlockOwnerDeletion,
		})
	}
}

// ownerToVirtualNode creates a virtual node as a placeholder for the Kubernetes owner object received in input.
// The virtual node will be eventually converted to an actual node when the node will be visited during discovery.
func (o *ObjectGraph) ownerToVirtualNode(owner metav1.OwnerReference, namespace string) *Node {
	isGlobal := false
	if namespace == "" {
		isGlobal = true
	}

	ownerNode := &Node{
		Identity: corev1.ObjectReference{
			APIVersion: owner.APIVersion,
			Kind:       owner.Kind,
			Name:       owner.Name,
			UID:        owner.UID,
			Namespace:  namespace,
		},
		Owners:         make(map[*Node]ownerReferenceAttributes),
		SoftOwners:     make(map[*Node]Empty),
		TenantClusters: make(map[*Node]Empty),
		TenantCRSs:     make(map[*Node]Empty),
		Virtual:        true,
		ForceMove:      o.getForceMove(owner.Kind, owner.APIVersion, nil),
		IsGlobal:       isGlobal,
	}

	o.uidToNode[ownerNode.Identity.UID] = ownerNode
	return ownerNode
}

// objToNode creates a node for the Kubernetes object received in input.
// If the node corresponding to the Kubernetes object already exists as a virtual node detected when processing OwnerReferences,
// the node is marked as Observed.
func (o *ObjectGraph) objToNode(obj *unstructured.Unstructured) *Node {
	existingNode, found := o.uidToNode[obj.GetUID()]
	if found {
		existingNode.markObserved()

		// In order to compensate the lack of labels when adding a virtual node,
		// it is required to re-compute the forceMove flag when the real node is processed
		// Without this, there is the risk that, forceMove will report false negatives depending on the discovery order
		existingNode.ForceMove = o.getForceMove(obj.GetKind(), obj.GetAPIVersion(), obj.GetLabels())
		return existingNode
	}

	isGlobal := false
	if obj.GetNamespace() == "" {
		isGlobal = true
	}

	newNode := &Node{
		Identity: corev1.ObjectReference{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			UID:        obj.GetUID(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
		},
		Owners:         make(map[*Node]ownerReferenceAttributes),
		SoftOwners:     make(map[*Node]Empty),
		TenantClusters: make(map[*Node]Empty),
		TenantCRSs:     make(map[*Node]Empty),
		Virtual:        false,
		ForceMove:      o.getForceMove(obj.GetKind(), obj.GetAPIVersion(), obj.GetLabels()),
		IsGlobal:       isGlobal,
	}

	o.uidToNode[newNode.Identity.UID] = newNode
	return newNode
}

func (o *ObjectGraph) getForceMove(kind, apiVersion string, labels map[string]string) bool {
	if _, ok := labels[meta.LabelUndistroMove]; ok {
		return true
	}

	kindAPIStr := getKindAPIString(metav1.TypeMeta{Kind: kind, APIVersion: apiVersion})

	if discoveryType, ok := o.types[kindAPIStr]; ok {
		return discoveryType.forceMove
	}
	return false
}

// getDiscoveryTypes returns the list of TypeMeta to be considered for the the move discovery phase.
// This list includes all the types defines by the CRDs installed by clusterctl and the ConfigMap/Secret core types.
func (o *ObjectGraph) GetDiscoveryTypes() error {
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
		return getCRDList(context.Background(), o.proxy, crdList)
	}); err != nil {
		return err
	}

	o.types = make(map[string]*discoveryTypeInfo)

	for _, crd := range crdList.Items {
		for _, version := range crd.Spec.Versions {
			if !version.Storage {
				continue
			}

			forceMove := false
			if _, ok := crd.Labels[meta.LabelUndistroMove]; ok {
				forceMove = true
			}

			typeMeta := metav1.TypeMeta{
				Kind: crd.Spec.Names.Kind,
				APIVersion: metav1.GroupVersion{
					Group:   crd.Spec.Group,
					Version: version.Name,
				}.String(),
			}

			o.types[getKindAPIString(typeMeta)] = &discoveryTypeInfo{
				typeMeta:  typeMeta,
				forceMove: forceMove,
			}

		}
	}

	secretTypeMeta := metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"}
	o.types[getKindAPIString(secretTypeMeta)] = &discoveryTypeInfo{typeMeta: secretTypeMeta}

	configMapTypeMeta := metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}
	o.types[getKindAPIString(configMapTypeMeta)] = &discoveryTypeInfo{typeMeta: configMapTypeMeta}

	return nil
}

// getKindAPIString returns a concatenated string of the API name and the plural of the kind
// Ex: KIND=Foo API NAME=foo.bar.domain.tld => foos.foo.bar.domain.tld
func getKindAPIString(typeMeta metav1.TypeMeta) string {
	api := strings.Split(typeMeta.APIVersion, "/")[0]
	return fmt.Sprintf("%ss.%s", strings.ToLower(typeMeta.Kind), api)
}

func getCRDList(ctx context.Context, c client.Client, crdList *apiextensionsv1.CustomResourceDefinitionList) error {
	if err := c.List(ctx, crdList, client.HasLabels{meta.LabelUndistro}); err != nil {
		return errors.Wrap(err, "failed to get the list of CRDs required for the move discovery phase")
	}
	return nil
}

// Discovery reads all the Kubernetes objects existing in a namespace (or in all namespaces if empty) for the types received in input, and then adds
// everything to the objects graph.
func (o *ObjectGraph) Discovery(namespace string) error {
	log := log.Log
	fmt.Fprintln(o.Out, "Discovering objects")

	selectors := []client.ListOption{
		client.InNamespace(namespace),
	}

	for _, discoveryType := range o.types {
		typeMeta := discoveryType.typeMeta
		objList := new(unstructured.UnstructuredList)

		if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return getObjList(o.proxy, typeMeta, selectors, objList)
		}); err != nil {
			return err
		}

		// if we are discovering Secrets, also secrets from the providers namespace should be included.
		if discoveryType.typeMeta.GetObjectKind().GroupVersionKind().GroupKind() == corev1.SchemeGroupVersion.WithKind("SecretList").GroupKind() {
			providers := configv1alpha1.ProviderList{}
			err := o.proxy.List(context.Background(), &providers)
			if err != nil {
				return err
			}
			for _, p := range providers.Items {
				if p.Spec.ProviderType == string(configv1alpha1.InfraProviderType) {
					providerNamespaceSelector := []client.ListOption{client.InNamespace(p.Namespace)}
					providerNamespaceSecretList := new(unstructured.UnstructuredList)
					if err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
						return getObjList(o.proxy, typeMeta, providerNamespaceSelector, providerNamespaceSecretList)
					}); err != nil {
						return err
					}
					objList.Items = append(objList.Items, providerNamespaceSecretList.Items...)
				}
			}
		}

		if len(objList.Items) == 0 {
			continue
		}

		log.V(5).Info(typeMeta.Kind, "Count", len(objList.Items))
		for i := range objList.Items {
			obj := objList.Items[i]
			o.addObj(&obj)
		}
	}

	log.V(1).Info("Total objects", "Count", len(o.uidToNode))

	// Completes the graph by searching for soft ownership relations such as secrets linked to the cluster
	// by a naming convention (without any explicit OwnerReference).
	o.setSoftOwnership()

	// Completes the graph by setting for each node the list of Clusters the node belong to.
	o.setClusterTenants()

	// Completes the graph by setting for each node the list of ClusterResourceSet the node belong to.
	o.setCRSTenants()

	return nil
}

func getObjList(c client.Client, typeMeta metav1.TypeMeta, selectors []client.ListOption, objList *unstructured.UnstructuredList) error {
	objList.SetAPIVersion(typeMeta.APIVersion)
	objList.SetKind(typeMeta.Kind)

	if err := c.List(context.Background(), objList, selectors...); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "failed to list %q resources", objList.GroupVersionKind())
	}
	return nil
}

// GetClusters returns the list of Clusters existing in the object graph.
func (o *ObjectGraph) GetClusters() []*Node {
	clusters := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.GroupVersionKind().GroupKind() == appv1alpha1.GroupVersion.WithKind("Cluster").GroupKind() && node.Identity.Name != "undistro-test" {
			clusters = append(clusters, node)
		}
	}
	return clusters
}

// GetCapiClusters returns the list of Clusters existing in the object graph.
func (o *ObjectGraph) GetCapiClusters() []*Node {
	clusters := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.GroupVersionKind().GroupKind() == clusterv1.GroupVersion.WithKind("Cluster").GroupKind() && node.Identity.Name != "capi-test" {
			clusters = append(clusters, node)
		}
	}
	return clusters
}

// getSecrets returns the list of Secrets existing in the object graph.
func (o *ObjectGraph) getSecrets() []*Node {
	secrets := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.APIVersion == "v1" && node.Identity.Kind == "Secret" {
			secrets = append(secrets, node)
		}
	}
	return secrets
}

// getNodes returns the list of nodes existing in the object graph.
func (o *ObjectGraph) getNodes() []*Node {
	nodes := []*Node{}
	for _, node := range o.uidToNode {
		nodes = append(nodes, node)
	}
	return nodes
}

// getCRSs returns the list of ClusterResourceSet existing in the object graph.
func (o *ObjectGraph) getCRSs() []*Node {
	clusters := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.GroupVersionKind().GroupKind() == addonsv1.GroupVersion.WithKind("ClusterResourceSet").GroupKind() {
			clusters = append(clusters, node)
		}
	}
	return clusters
}

// GetMoveNodes returns the list of nodes existing in the object graph that belong at least to one Cluster or to a ClusterResourceSet
// or to a CRD containing the "move" label.
func (o *ObjectGraph) GetMoveNodes() []*Node {
	nodes := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.GroupVersionKind().GroupKind() == appv1alpha1.GroupVersion.WithKind("Cluster").GroupKind() && node.Identity.Name != "undistro-test" {
			nodes = append(nodes, node)
			continue
		}
		if len(node.TenantClusters) > 0 || len(node.TenantCRSs) > 0 || node.ForceMove {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// getMachines returns the list of Machine existing in the object graph.
func (o *ObjectGraph) GetMachines() []*Node {
	machines := []*Node{}
	for _, node := range o.uidToNode {
		if node.Identity.GroupVersionKind().GroupKind() == clusterv1.GroupVersion.WithKind("Machine").GroupKind() {
			machines = append(machines, node)
		}
	}
	return machines
}

// setSoftOwnership searches for soft ownership relations such as secrets linked to the cluster by a naming convention (without any explicit OwnerReference).
func (o *ObjectGraph) setSoftOwnership() {
	clusters := o.GetCapiClusters()
	for _, secret := range o.getSecrets() {
		// If the secret has at least one OwnerReference ignore it.
		// NB. Cluster API generated secrets have an explicit OwnerReference to the ControlPlane or the KubeadmConfig object while user provided secrets might not have one.
		if len(secret.Owners) > 0 {
			continue
		}

		// If the secret is linked to a cluster, then add the cluster to the list of the secrets's softOwners.
		for _, cluster := range clusters {
			secret.addSoftOwner(cluster)
		}
	}
}

// setClusterTenants sets the cluster tenants for the clusters itself and all their dependent object tree.
func (o *ObjectGraph) setClusterTenants() {
	for _, cluster := range o.GetCapiClusters() {
		o.setClusterTenant(cluster, cluster)
	}
}

// setNodeTenant sets a cluster tenant for a node and for its own dependents/sofDependents.
func (o *ObjectGraph) setClusterTenant(node, tenant *Node) {
	node.TenantClusters[tenant] = Empty{}
	for _, other := range o.getNodes() {
		if other.isOwnedBy(node) || other.isSoftOwnedBy(node) {
			o.setClusterTenant(other, tenant)
		}
	}
}

// setClusterTenants sets the ClusterResourceSet tenants for the ClusterResourceSet itself and all their dependent object tree.
func (o *ObjectGraph) setCRSTenants() {
	for _, crs := range o.getCRSs() {
		o.setCRSTenant(crs, crs)
	}
}

// setCRSTenant sets a ClusterResourceSet tenant for a node and for its own dependents/sofDependents.
func (o *ObjectGraph) setCRSTenant(node, tenant *Node) {
	node.TenantCRSs[tenant] = Empty{}
	for _, other := range o.getNodes() {
		if other.isOwnedBy(node) {
			o.setCRSTenant(other, tenant)
		}
	}
}

// checkVirtualNode logs if nodes are still virtual
func (o *ObjectGraph) CheckVirtualNode() {
	log := log.Log
	for _, node := range o.uidToNode {
		if node.Virtual {
			log.V(5).Info("Object won't be moved because it's not included in GVK considered for move", "kind", node.Identity.Kind, "name", node.Identity.Name)
		}
	}
}

// MoveSequence defines a list of group of moveGroups
type MoveSequence struct {
	Groups   []MoveGroup
	NodesMap map[*Node]Empty
}

// MoveGroup defines is a list of nodes read from the object graph that can be moved in parallel.
type MoveGroup []*Node

func (s *MoveSequence) AddGroup(group MoveGroup) {
	// Add the group
	s.Groups = append(s.Groups, group)
	// Add all the nodes in the group to the nodeMap so we can check if a node is already in the move sequence or not
	for _, n := range group {
		s.NodesMap[n] = Empty{}
	}
}

func (s *MoveSequence) HasNode(n *Node) bool {
	_, ok := s.NodesMap[n]
	return ok
}

func (s *MoveSequence) GetGroup(i int) MoveGroup {
	return s.Groups[i]
}

// Define the move sequence by processing the ownerReference chain.
func GetMoveSequence(graph *ObjectGraph) *MoveSequence {
	moveSequence := &MoveSequence{
		Groups:   []MoveGroup{},
		NodesMap: make(map[*Node]Empty),
	}

	for {
		// Determine the next move group by processing all the nodes in the graph that belong to a Cluster.
		// NB. it is necessary to filter out nodes not belonging to a cluster because e.g. discovery reads all the secrets,
		// but only few of them are related to Clusters/Machines etc.
		moveGroup := MoveGroup{}

		for _, n := range graph.GetMoveNodes() {
			// If the node was already included in the moveSequence, skip it.
			if moveSequence.HasNode(n) {
				continue
			}

			// Check if all the ownerReferences are already included in the move sequence; if yes, add the node to move group,
			// otherwise skip it (the node will be re-processed in the next group).
			ownersInPlace := true
			for owner := range n.Owners {
				if !moveSequence.HasNode(owner) {
					ownersInPlace = false
					break
				}
			}
			for owner := range n.SoftOwners {
				if !moveSequence.HasNode(owner) {
					ownersInPlace = false
					break
				}
			}
			if ownersInPlace {
				moveGroup = append(moveGroup, n)
			}
		}

		// If the resulting move group is empty it means that all the nodes are already in the sequence, so exit.
		if len(moveGroup) == 0 {
			break
		}
		moveSequence.AddGroup(moveGroup)
	}
	return moveSequence
}
