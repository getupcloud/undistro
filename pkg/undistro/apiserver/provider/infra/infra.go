/*
Copyright 2021 The UnDistro authors

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
package infra

import (
	"errors"

	typesv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra/aws"
	"k8s.io/client-go/rest"
)

var ErrInvalidProviderName = errors.New("a valid infra provider name is required.\n" +
	"supported are ['aws']")

type ProviderParams struct {
	Name         string       `json:"name"`
	Region       string       `json:"region,omitempty"`
	Meta         string       `json:"meta,omitempty"`
	Page         int          `json:"page,omitempty"`
	ItemsPerPage int          `json:"items_per_page,omitempty"`
	Config       *rest.Config `json:"rest_config,omitempty"`
}

func New(conf *rest.Config, name, region, meta string, page, itemsPerPage int) *ProviderParams {
	return &ProviderParams{
		Name:         name,
		Region:       region,
		Meta:         meta,
		Page:         page,
		ItemsPerPage: itemsPerPage,
		Config:       conf,
	}
}

func (pr *ProviderParams) DescribeMetadata() (result interface{}, err error) {
	switch pr.Name {
	case typesv1alpha1.Amazon.String():
		result, err = aws.DescribeMeta(pr.Config, pr.Region, pr.Meta, pr.Page, pr.ItemsPerPage)
		if err != nil {
			return nil, err
		}
		return
	}
	return nil, ErrInvalidProviderName
}
