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

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/util"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
)

type provider struct {
	Name         string
	ProviderType string
}

type test struct {
	name           string
	params         provider
	expectedStatus int
	error          error
	body           interface{}
}

func TestRetrieveMetadata(t *testing.T) {
	cases := []test{
		{
			name: "test get metadata passing invalid provider",
			params: provider{
				Name:         "amazon",
				ProviderType: string(configv1alpha1.InfraProviderType),
			},
			expectedStatus: http.StatusBadRequest,
			error:          infra.ErrInvalidProvider,
		},
		{
			name: "test get metadata passing no provider",
			params: provider{
				Name:         "_",
				ProviderType: string(configv1alpha1.InfraProviderType),
			},
			expectedStatus: http.StatusBadRequest,
			error:          infra.ErrInvalidProvider,
		},
		{
			name: "test get metadata passing provider wrong type",
			params: provider{
				Name:         appv1alpha1.Amazon.String(),
				ProviderType: string(configv1alpha1.CoreProviderType),
			},
			expectedStatus: http.StatusBadRequest,
			error:          errReadQueryParam,
		},
		{
			name: "test successfully infra provider metadata",
			params: provider{
				Name:         appv1alpha1.Amazon.String(),
				ProviderType: string(configv1alpha1.InfraProviderType),
			},
			expectedStatus: http.StatusOK,
			error:          nil,
		},
	}

	h := Handler{DefaultConfig: nil}

	r := mux.NewRouter()
	r.HandleFunc("/provider/{name}/metadata", h.HandleProviderMetadata)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, p := range cases {
		t.Run(p.name, func(t *testing.T) {
			endpoint := fmt.Sprintf("%s/provider/%s/metadata", ts.URL, p.params.Name)

			req, err := http.NewRequest(http.MethodGet, endpoint, nil)
			if err != nil {
				t.Errorf("error: %s\n", err.Error())
			}
			// add provider type
			q := req.URL.Query()
			q.Add("provider_type", p.params.ProviderType)
			req.URL.RawQuery = q.Encode()

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("error: %s\n", err.Error())
			}

			if status := resp.StatusCode; status != p.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v\n",
					status, p.expectedStatus)
			}

			// validate body
			var received util.ErrResponder
			if p.error != nil {
				byt, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("error: %s\n", err.Error())
				}

				err = json.Unmarshal(byt, &received)

				if err != nil {
					t.Errorf("error: %s\n", err.Error())
				}

				if received.Message != p.error.Error() {
					t.Errorf("handler returned unexpected body: got %v want %v",
						received.Message, p.error.Error())
				}
			}
		})
	}
}
