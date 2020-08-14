/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package config

import (
	"encoding/json"
	"path/filepath"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
)

// Provider defines a provider configuration.
type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// Type returns the type of the provider.
	Type() undistrov1.ProviderType

	// URL returns the name of the provider repository.
	URL() string

	// SameAs returns true if two providers have the same name and type.
	// Please note that this uniquely identifies a provider configuration, but not the provider instances in the cluster
	// because it is possible to create many instances of the same provider.
	SameAs(other Provider) bool

	// ManifestLabel returns the cluster.x-k8s.io/provider label value for a provider.
	// Please note that this label uniquely identifies the provider, e.g. bootstrap-kubeadm, but not the instances of
	// the provider, e.g. namespace-1/bootstrap-kubeadm and namespace-2/bootstrap-kubeadm
	ManifestLabel() string

	// Less func can be used to ensure a consist order of provider lists.
	Less(other Provider) bool

	// GetInitFunc return provider initFunc
	GetInitFunc() InitFunc

	// GetPreConfigFunc return preConfigFunc
	GetPreConfigFunc() PreConfigFunc
}

type InitFunc func(Client, bool) error

type PreConfigFunc func(*undistrov1.Cluster, VariablesClient) error

// provider implements provider
type provider struct {
	name          string
	url           string
	providerType  undistrov1.ProviderType
	initFunc      InitFunc
	preConfigFunc PreConfigFunc
}

// ensure provider implements provider
var _ Provider = &provider{}

func (p *provider) Name() string {
	return p.name
}

func (p *provider) URL() string {
	return p.url
}

func (p *provider) Type() undistrov1.ProviderType {
	return p.providerType
}

func (p *provider) SameAs(other Provider) bool {
	return p.name == other.Name() && p.providerType == other.Type()
}

func (p *provider) ManifestLabel() string {
	return undistrov1.ManifestLabel(p.name, p.Type())
}

func (p *provider) Less(other Provider) bool {
	return p.providerType.Order() < other.Type().Order() ||
		(p.providerType.Order() == other.Type().Order() && p.name < other.Name())
}

func (p *provider) GetInitFunc() InitFunc {
	return p.initFunc
}

func (p *provider) GetPreConfigFunc() PreConfigFunc {
	return p.preConfigFunc
}

func NewProvider(name string, url string, ttype undistrov1.ProviderType, initFunc InitFunc, preConfigFunc PreConfigFunc) Provider {
	return &provider{
		name:          name,
		url:           url,
		providerType:  ttype,
		initFunc:      initFunc,
		preConfigFunc: preConfigFunc,
	}
}

func (p provider) MarshalJSON() ([]byte, error) {
	dir, file := filepath.Split(p.url)
	j, err := json.Marshal(struct {
		Name         string
		ProviderType undistrov1.ProviderType
		URL          string
		File         string
	}{
		Name:         p.name,
		ProviderType: p.providerType,
		URL:          dir,
		File:         file,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}
