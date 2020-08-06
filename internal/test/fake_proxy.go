/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package test

import (
	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	fakebootstrap "github.com/getupcloud/undistro/internal/test/providers/bootstrap"
	fakecontrolplane "github.com/getupcloud/undistro/internal/test/providers/controlplane"
	fakeinfrastructure "github.com/getupcloud/undistro/internal/test/providers/infrastructure"
	apiextensionslv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	addonsv1alpha3 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha3"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type FakeProxy struct {
	cs   client.Client
	objs []runtime.Object
}

var (
	FakeScheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(FakeScheme)
	_ = undistrov1.AddToScheme(FakeScheme)
	_ = clusterv1.AddToScheme(FakeScheme)
	_ = expv1.AddToScheme(FakeScheme)
	_ = addonsv1alpha3.AddToScheme(FakeScheme)
	_ = apiextensionslv1.AddToScheme(FakeScheme)

	_ = fakebootstrap.AddToScheme(FakeScheme)
	_ = fakecontrolplane.AddToScheme(FakeScheme)
	_ = fakeinfrastructure.AddToScheme(FakeScheme)
}

func (f *FakeProxy) CurrentNamespace() (string, error) {
	return "default", nil
}

func (f *FakeProxy) ValidateKubernetesVersion() error {
	return nil
}

func (f *FakeProxy) GetConfig() (*rest.Config, error) {
	return nil, nil
}

func (f *FakeProxy) NewClient() (client.Client, error) {
	if f.cs != nil {
		return f.cs, nil
	}
	f.cs = fake.NewFakeClientWithScheme(FakeScheme, f.objs...)

	return f.cs, nil
}

// ListResources returns all the resources known by the FakeProxy
func (f *FakeProxy) ListResources(labels map[string]string, namespaces ...string) ([]unstructured.Unstructured, error) {
	var ret []unstructured.Unstructured //nolint
	for _, o := range f.objs {
		u := unstructured.Unstructured{}
		err := FakeScheme.Convert(o, &u, nil)
		if err != nil {
			return nil, err
		}

		// filter by namespace, if any
		if len(namespaces) > 0 && u.GetNamespace() != "" {
			inNamespaces := false
			for _, namespace := range namespaces {
				if u.GetNamespace() == namespace {
					inNamespaces = true
					break
				}
			}
			if !inNamespaces {
				continue
			}
		}

		// filter by label, if any
		haslabel := false
		for l, v := range labels {
			for ul, uv := range u.GetLabels() {
				if l == ul && v == uv {
					haslabel = true
				}
			}
		}
		if !haslabel {
			continue
		}

		ret = append(ret, u)
	}

	return ret, nil
}

func NewFakeProxy() *FakeProxy {
	return &FakeProxy{}
}

func (f *FakeProxy) WithObjs(objs ...runtime.Object) *FakeProxy {
	f.objs = append(f.objs, objs...)
	return f
}

// WithProviderInventory can be used as a fast track for setting up test scenarios requiring an already initialized management cluster.
// NB. this method adds an items to the Provider inventory, but it doesn't install the corresponding provider; if the
// test case requires the actual provider to be installed, use the the fake client to install both the provider
// components and the corresponding inventory item.
func (f *FakeProxy) WithProviderInventory(name string, providerType undistrov1.ProviderType, version, targetNamespace, watchingNamespace string) *FakeProxy {
	f.objs = append(f.objs, &undistrov1.Provider{
		TypeMeta: metav1.TypeMeta{
			APIVersion: undistrov1.GroupVersion.String(),
			Kind:       "Provider",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: targetNamespace,
			Name:      undistrov1.ManifestLabel(name, providerType),
			Labels: map[string]string{
				undistrov1.ClusterctlLabelName:     "",
				undistrov1.UndistroLabelName:       "",
				clusterv1.ProviderLabelName:        undistrov1.ManifestLabel(name, providerType),
				undistrov1.ClusterctlCoreLabelName: "inventory",
				undistrov1.UndistroCoreLabelName:   "inventory",
			},
		},
		ProviderName:     name,
		Type:             string(providerType),
		Version:          version,
		WatchedNamespace: watchingNamespace,
	})

	return f
}
