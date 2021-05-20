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
	"k8s.io/klog/v2"
	"net/http"

	"github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/gorilla/mux"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

type Metadata struct {
	Type v1alpha1.InfrastructureProvider
	MachineTypes []string `json:"machine_types"`
	ProviderRegions []string `json:"provider_regions"`
	SupportedFlavors []string `json:"supported_flavors"` // k8s flavors (ec2,eks, etc) and each version
}

type ErrResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var (
	NoProviderName  = errors.New("no provider name was found")
	ReadQueryParam  = errors.New("query param invalid")
	InvalidProvider = errors.New("invalid provider, maybe unsupported")
)

// ServeHTTP retrieves infraProviderMetadata about some Provider
func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	vars := mux.Vars(r)
	pn := vars["name"]
	if pn == "" {
		http.Error(w, NoProviderName.Error(), http.StatusBadRequest)
		return
	}

	// extract provider type
	providerType := r.URL.Query().Get("provider_type")
	if providerType == "" {
		http.Error(w, ReadQueryParam.Error(), http.StatusInternalServerError)
		return
	}

	switch providerType {
	case string(v1alpha3.InfrastructureProviderType):
		if !isValidInfraProvider(pn) {
			http.Error(w, InvalidProvider.Error(), http.StatusBadRequest)
			return
		}
		infraProviderMetadata(pn, w)
	default:
		// invalid provider type
		http.Error(w, ReadQueryParam.Error(), http.StatusBadRequest)
	}
}

func infraProviderMetadata(providerName string, w http.ResponseWriter) {
	//generate Infrastructure Provider Metadata info about the provider
	pm := Metadata{
		Type:             v1alpha1.InfrastructureProvider{},
		MachineTypes:     nil,
		ProviderRegions:  nil,
		SupportedFlavors: nil,
	}
	klog.Infoln(pm)

	w.WriteHeader(http.StatusOK)
}

func isValidInfraProvider(name string) bool {
	return name == v1alpha1.Amazon.String()
}
