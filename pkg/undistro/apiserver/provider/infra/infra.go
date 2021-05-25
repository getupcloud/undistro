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

func IsValidInfraProvider(p string) bool {
	return p == typesv1alpha1.Amazon.String()
}

var ErrInvalidProviderName = errors.New("name is required. supported are ['aws']")

func DescribeInfraMetadata(config *rest.Config, name, meta string, page int) (result interface{}, err error) {
	switch name {
	case typesv1alpha1.Amazon.String():
		result, err = aws.DescribeMeta(config, meta, page)
		if err != nil {
			return nil, err
		}
		return
	}
	return nil, ErrInvalidProviderName
}


