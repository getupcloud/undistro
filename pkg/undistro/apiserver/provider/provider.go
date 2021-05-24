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
	"errors"
	"net/http"

	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/gorilla/mux"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

var (
	errNoProviderName = errors.New("no provider name was found")
	readQueryParam    = errors.New("query param invalid")
)

// MetadataHandler retrieves Provider Metadata
func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	vars := mux.Vars(r)
	pn := vars["name"]
	if pn == "" {
		http.Error(w, errNoProviderName.Error(), http.StatusBadRequest)
		return
	}

	// extract provider type
	providerType := r.URL.Query().Get("provider_type")
	if providerType == "" {
		providerType = string(v1alpha3.CoreProviderType)
	}

	// write metadata by provider type
	switch providerType {
	case string(v1alpha3.InfrastructureProviderType):
		infra.WriteMetadata(pn, w)
	default:
		// invalid provider type
		http.Error(w, readQueryParam.Error(), http.StatusBadRequest)
	}
}
