/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package scheme

import (
	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	addonsv1alpha3 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
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
	_ = addonsv1alpha3.AddToScheme(Scheme)
	// +kubebuilder:scaffold:scheme
}
