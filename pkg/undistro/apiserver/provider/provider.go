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

	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/util"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
)

var (
	errNoProviderName = errors.New("no provider name was found")
	readQueryParam    = errors.New("query param invalid or empty")
)

type Handler struct {
	DefaultConfig *rest.Config
}

// MetadataHandler retrieves Provider metadata
func (h *Handler) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	vars := mux.Vars(r)
	pn := vars["name"]
	if pn == "" {
		util.WriteError(w, errNoProviderName, http.StatusBadRequest)
		return
	}

	// extract provider type
	providerType := r.URL.Query().Get("provider_type")
	if providerType == "" {
		providerType = string(configv1alpha1.InfraProviderType)
	}

	// write metadata by provider type
	switch providerType {
	case string(configv1alpha1.InfraProviderType):
		infra.WriteMetadata(pn, w)
	default:
		// invalid provider type
		util.WriteError(w, readQueryParam, http.StatusBadRequest)
	}
}

func (h Handler) SSHKeysHandler(w http.ResponseWriter, r *http.Request) {
	// extract region
	region := r.URL.Query().Get("region")
	if region == "" {
		util.WriteError(w, readQueryParam, http.StatusBadRequest)
	}

	// retrieve ssh keys
	keys, err := infra.DescribeSSHKeys(region, h.DefaultConfig)
	if err != nil {
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	// write response body
	encoder := json.NewEncoder(w)
	err = encoder.Encode(keys)
	if err != nil {
		util.WriteError(w, err, http.StatusInternalServerError)
		return
	}
}
