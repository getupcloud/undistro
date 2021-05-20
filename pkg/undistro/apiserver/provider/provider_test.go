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
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

type provider struct {
	Name         string `json:"name"`
	ProviderType string `json:"providerType"`
}

type test struct {
	name           string
	params         provider
	expectedStatus int
}

func TestRetrieveMetadata(t *testing.T) {
	cases := []test{
		{
			name: "test get metadata passing invalid provider",
			params: provider{
				Name:         "amazon",
				ProviderType: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "test get metadata passing no provider",
			params: provider{
				Name:         " ",
				ProviderType: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "test get metadata passing provider wrong type",
			params: provider{
				Name:         v1alpha1.Amazon.String(),
				ProviderType: string(v1alpha3.CoreProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "test successfully infra provider metadata",
			params: provider{
				Name:         v1alpha1.Amazon.String(),
				ProviderType: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusOK,
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/provider/{name}/metadata", MetadataHandler)

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
		})
	}
}

// validate metadata body
//expected := `{"alive": true}`
//if rr.Body.String() != expected {
//	t.Errorf("handler returned unexpected body: got %v want %v",
//		rr.Body.String(), expected)
//}
