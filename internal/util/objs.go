/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util

import (
	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/internal/scheme"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	deploymentKind          = "Deployment"
	controllerContainerName = "manager"
)

// InspectImages identifies the container images required to install the objects defined in the objs.
// NB. The implemented approach is specific for the provider components YAML & for the cert-manager manifest; it is not
// intended to cover all the possible objects used to deploy containers existing in Kubernetes.
func InspectImages(objs []unstructured.Unstructured) ([]string, error) {
	images := []string{}

	for i := range objs {
		o := objs[i]
		if o.GetKind() == deploymentKind {
			d := &appsv1.Deployment{}
			if err := scheme.Scheme.Convert(&o, d, nil); err != nil {
				return nil, err
			}

			for _, c := range d.Spec.Template.Spec.Containers {
				images = append(images, c.Image)
			}

			for _, c := range d.Spec.Template.Spec.InitContainers {
				images = append(images, c.Image)
			}
		}
	}

	return images, nil
}

// IsClusterResource returns true if the resource kind is cluster wide (not namespaced).
func IsClusterResource(kind string) bool {
	return !IsResourceNamespaced(kind)
}

// IsResourceNamespaced returns true if the resource kind is namespaced.
func IsResourceNamespaced(kind string) bool {
	switch kind {
	case "Namespace",
		"Node",
		"PersistentVolume",
		"PodSecurityPolicy",
		"CertificateSigningRequest",
		"ClusterRoleBinding",
		"ClusterRole",
		"VolumeAttachment",
		"StorageClass",
		"CSIDriver",
		"CSINode",
		"ValidatingWebhookConfiguration",
		"MutatingWebhookConfiguration",
		"CustomResourceDefinition",
		"PriorityClass",
		"RuntimeClass":
		return false
	default:
		return true
	}
}

// IsSharedResource returns true if the resource lifecycle is shared.
func IsSharedResource(o unstructured.Unstructured) bool {
	labels := o.GetLabels()
	lifecycle := labels[undistrov1.ClusterctlResourceLifecyleLabelName]
	lifecycleUndistro := labels[undistrov1.UndistroResourceLifecyleLabelName]
	exp := string(undistrov1.ResourceLifecycleShared)
	if lifecycle == exp || lifecycleUndistro == exp {
		return true
	}
	return false
}

// FixImages alters images using the give alter func
// NB. The implemented approach is specific for the provider components YAML & for the cert-manager manifest; it is not
// intended to cover all the possible objects used to deploy containers existing in Kubernetes.
func FixImages(objs []unstructured.Unstructured, alterImageFunc func(image string) (string, error)) ([]unstructured.Unstructured, error) {
	// look for resources of kind Deployment and alter the image
	for i := range objs {
		o := &objs[i]
		if o.GetKind() != deploymentKind {
			continue
		}

		// Convert Unstructured into a typed object
		d := &appsv1.Deployment{}
		if err := scheme.Scheme.Convert(o, d, nil); err != nil {
			return nil, err
		}

		// Alter the image
		for j := range d.Spec.Template.Spec.Containers {
			container := d.Spec.Template.Spec.Containers[j]
			image, err := alterImageFunc(container.Image)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to fix image for container %s in deployment %s", container.Name, d.Name)
			}
			container.Image = image
			d.Spec.Template.Spec.Containers[j] = container
		}

		for j := range d.Spec.Template.Spec.InitContainers {
			container := d.Spec.Template.Spec.InitContainers[j]
			image, err := alterImageFunc(container.Image)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to fix image for init container %s in deployment %s", container.Name, d.Name)
			}
			container.Image = image
			d.Spec.Template.Spec.InitContainers[j] = container
		}

		// Convert typed object back to Unstructured
		if err := scheme.Scheme.Convert(d, o, nil); err != nil {
			return nil, err
		}
		objs[i] = *o
	}
	return objs, nil
}

func ReverseObjs(s []unstructured.Unstructured) []unstructured.Unstructured {
	a := make([]unstructured.Unstructured, len(s))
	copy(a, s)
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
	return a
}
