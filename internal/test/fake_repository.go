/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package test

import (
	"fmt"

	clusterctlv1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type FakeRepository struct {
	defaultVersion string
	rootPath       string
	componentsPath string
	versions       map[string]bool
	files          map[string][]byte
}

func (f *FakeRepository) DefaultVersion() string {
	return f.defaultVersion
}

func (f *FakeRepository) RootPath() string {
	return f.rootPath
}

func (f *FakeRepository) ComponentsPath() string {
	return f.componentsPath
}

func (f FakeRepository) GetFile(version string, path string) ([]byte, error) {
	if _, ok := f.versions[version]; !ok {
		return nil, errors.Errorf("unable to get files for version %s", version)
	}

	for p, c := range f.files {
		if p == vpath(version, path) {
			return c, nil
		}
	}
	return nil, errors.Errorf("unable to get file %s for version %s", path, version)
}

func (f *FakeRepository) GetVersions() ([]string, error) {
	v := make([]string, 0, len(f.versions))
	for k := range f.versions {
		v = append(v, k)
	}
	return v, nil
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		versions: map[string]bool{},
		files:    map[string][]byte{},
	}
}

func (f *FakeRepository) WithPaths(rootPath, componentsPath string) *FakeRepository {
	f.rootPath = rootPath
	f.componentsPath = componentsPath
	return f
}

func (f *FakeRepository) WithDefaultVersion(version string) *FakeRepository {
	f.defaultVersion = version
	return f
}

func (f *FakeRepository) WithFile(version, path string, content []byte) *FakeRepository {
	f.versions[version] = true
	f.files[vpath(version, path)] = content
	return f
}

func (f *FakeRepository) WithVersions(version ...string) *FakeRepository {
	for _, v := range version {
		f.versions[v] = true
	}
	return f
}

func (f *FakeRepository) WithMetadata(version string, metadata *clusterctlv1.Metadata) *FakeRepository {
	scheme := runtime.NewScheme()
	if err := clusterctlv1.AddToScheme(scheme); err != nil {
		panic(err)
	}

	codecs := serializer.NewCodecFactory(scheme)

	mediaType := "application/yaml"
	info, match := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !match {
		panic("failed to get SerializerInfo for application/yaml")
	}

	metadata.SetGroupVersionKind(clusterctlv1.GroupVersion.WithKind("Metadata"))

	encoder := codecs.EncoderForVersion(info.Serializer, metadata.GroupVersionKind().GroupVersion())
	data, err := runtime.Encode(encoder, metadata)
	if err != nil {
		panic(err)
	}

	return f.WithFile(version, "metadata.yaml", data)
}

func vpath(version string, path string) string {
	return fmt.Sprintf("%s/%s", version, path)
}
