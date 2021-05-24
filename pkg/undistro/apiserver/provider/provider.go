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
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
	"net/http"
)

var (
	errNoProviderName = errors.New("no provider name was found")
	readQueryParam    = errors.New("query param invalid or empty")
)

type ErrResponder struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Handler struct {
	DefaultConfig *rest.Config
}

// MetadataHandler retrieves Provider metadata
func (h *Handler) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	vars := mux.Vars(r)
	pn := vars["name"]
	if pn == "" {
		writeError(w, errNoProviderName, http.StatusBadRequest)
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
		writeError(w, readQueryParam, http.StatusBadRequest)
	}
}

func (h Handler) SSHKeysHandler(w http.ResponseWriter, r *http.Request) {
	// extract region
	region := r.URL.Query().Get("region")
	if region == "" {
		writeError(w, readQueryParam, http.StatusBadRequest)
	}

	res, err := infra.DescribeSSHKeys(region, h.DefaultConfig)

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(res)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}
}

func writeError(w http.ResponseWriter, err error, code int) {
	resp := ErrResponder{
		Status:  http.StatusText(code),
		Code:    code,
		Message: err.Error(),
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
