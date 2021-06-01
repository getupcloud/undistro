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
package provider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	apisv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra/aws"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestRetrieveMetadata(t *testing.T) {
	cases := []struct {
		name           string
		params         map[string]string
		expectedStatus int
		expectedErr    error
		body           interface{}
	}{
		{
			name: "test get metadata passing no provider name",
			params: map[string]string{
				ParamName: "",
				ParamType: string(configv1alpha1.InfraProviderType),
				ParamMeta: string(aws.RegionsMeta),
			},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    errEmptyProviderName,
		},
		{
			name: "test get metadata passing unsupported provider",
			params: map[string]string{
				ParamName: "amazon",
				ParamType: string(configv1alpha1.InfraProviderType),
				ParamMeta: string(aws.RegionsMeta),
			},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    infra.ErrInvalidProviderName,
		},
		{
			name: "test get metadata using default provider type",
			params: map[string]string{
				ParamName: apisv1alpha1.Amazon.String(),
				ParamMeta: string(aws.RegionsMeta),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "test get metadata with unsupported provider type",
			params: map[string]string{
				ParamName: apisv1alpha1.Amazon.String(),
				ParamMeta: string(aws.RegionsMeta),
				ParamType: string(configv1alpha1.CoreProviderType),
			},
			expectedStatus: http.StatusBadRequest,
			expectedErr:   errCoreProviderNotSupported,
		},
		{
			name: "test get metadata without meta param",
			params: map[string]string{
				ParamName: apisv1alpha1.Amazon.String(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedErr: aws.ErrNoProviderMeta,
		},
	}

	h := Handler{DefaultConfig: nil}

	r := mux.NewRouter()
	r.HandleFunc("/provider/metadata", h.HandleProviderMetadata)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			endpoint := fmt.Sprintf("%s/provider/metadata", ts.URL)

			req, err := http.NewRequest(http.MethodGet, endpoint, nil)
			if err != nil {
				t.Errorf("error: %s\n", err.Error())
			}

			// add params
			q := req.URL.Query()

			for k, v := range test.params {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("error: %s\n", err.Error())
			}

			if status := resp.StatusCode; status != test.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v\n",
					status, test.expectedStatus)
			}

			// validate error body
			var actual errResponse
			if test.expectedErr != nil {
				byt, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("error: %s\n", err.Error())
				}

				err = json.Unmarshal(byt, &actual)

				if err != nil {
					t.Errorf("error: %s\n", err.Error())
				}

				if actual.Message != test.expectedErr.Error() {
					t.Errorf("handler returned unexpected body: got %v want %v",
						actual.Message, test.expectedErr.Error())
				}
			}
		})
	}
}
