/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package scheme

import (
	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	admission "k8s.io/api/admission/v1"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	awsv1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	awsbotstrapv1 "sigs.k8s.io/cluster-api-provider-aws/bootstrap/eks/api/v1alpha3"
	awscontrolplanev1 "sigs.k8s.io/cluster-api-provider-aws/controlplane/eks/api/v1alpha3"
	awsexpinfrav1 "sigs.k8s.io/cluster-api-provider-aws/exp/api/v1alpha3"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	kuneadmbv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	kubeadmcpv1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	addonsv1alpha3 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
	expv1alpha3 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
)

var (
	// Scheme contains a set of API resources used by undistro
	Scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = undistrov1.AddToScheme(Scheme)
	_ = clusterv1.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = apiextensionsv1beta1.AddToScheme(Scheme)
	_ = addonsv1alpha3.AddToScheme(Scheme)
	_ = kubeadmcpv1.AddToScheme(Scheme)
	_ = awsv1.AddToScheme(Scheme)
	_ = admissionregistrationv1beta1.AddToScheme(Scheme)
	_ = admissionregistration.AddToScheme(Scheme)
	_ = expv1alpha3.AddToScheme(Scheme)
	_ = addonsv1alpha3.AddToScheme(Scheme)
	_ = awsbotstrapv1.AddToScheme(Scheme)
	_ = awscontrolplanev1.AddToScheme(Scheme)
	_ = awsexpinfrav1.AddToScheme(Scheme)
	_ = vspherev1.AddToScheme(Scheme)
	_ = kuneadmbv1.AddToScheme(Scheme)
	_ = admission.AddToScheme(Scheme)
	// +kubebuilder:scaffold:scheme
}
