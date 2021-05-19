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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"testing"
)

type params struct {
	providerName string
	providerKind string
}

type test struct {
	name 	string
	params params
	expectedStatus int
}

func TestRetrieveMetadata(t *testing.T) {
	cases := []test{
		{
			name:           "test get metadata passing invalid provider",
			params:      params{
				providerName: "amazon",
				providerKind: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test get metadata passing no provider",
			params:       params{
				providerName: "",
				providerKind: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test get metadata passing provider wrong type",
			params:         params{
				providerName: "aws",
				providerKind: string(v1alpha3.CoreProviderType),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test successfully infra provider metadata",
			params:         params{
				providerName: v1alpha1.Aws.String(),
				providerKind: string(v1alpha3.InfrastructureProviderType),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, p := range cases {
		t.Run(p.name, func(t *testing.T) {
			var endpoint = fmt.Sprintf("/provider/%s/metadata", p.params.providerName)

			body := struct {
				ProviderKind string `json:"providerKind"`
			}{
				ProviderKind: p.params.providerKind,
			}
			byt, err := json.Marshal(&body)
			if err != nil {
				t.Fatalf("error: %s\n", err.Error())
			}
			req, err := http.NewRequest(http.MethodGet, endpoint, bytes.NewBuffer(byt))
			if err != nil {
				t.Fatalf("error: %s\n", err.Error())
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(RetrieveMetadata)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != p.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, p.expectedStatus)
			}
			// validate body
			//expected := `{"alive": true}`
			//if rr.Body.String() != expected {
			//	t.Errorf("handler returned unexpected body: got %v want %v",
			//		rr.Body.String(), expected)
			//}
		})
	}
}
