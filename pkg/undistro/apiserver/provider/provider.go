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
	"strconv"

	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra"
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
)

var (
	errNoProviderName = errors.New("no provider name was found")
	errReadQueryParam = errors.New("query param invalid or empty")
	errInvalidPageRange = errors.New("invalid page range")
)

type Handler struct {
	DefaultConfig *rest.Config
}

func NewHandler(cfg *rest.Config) *Handler {
	return &Handler{
		DefaultConfig: cfg,
	}
}

// HandleProviderMetadata retrieves Provider metadata  by type
func (h *Handler) HandleProviderMetadata(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	pn := routeField(r, "name")
	if pn == "" {
		writeError(w, errNoProviderName, http.StatusBadRequest)
		return
	}

	// extract provider type
	providerType := queryField(r, "provider_type")
	if providerType == "" {
		providerType = string(configv1alpha1.InfraProviderType)
	}

	// write metadata by provider type
	switch providerType {
	case string(configv1alpha1.InfraProviderType):
		meta, err := infra.DescribeMetadata(pn)
		if err != nil {
			writeError(w, infra.ErrInvalidProvider, http.StatusBadRequest)
			return
		}
		writeResponse(w, meta)
		return
	default:
		// invalid provider type
		writeError(w, errReadQueryParam, http.StatusBadRequest)
	}
}

// HandleMachineTypes receives an integer page value and returns 10 items
func (h *Handler) HandleMachineTypes(w http.ResponseWriter, r *http.Request) {
	// extract provider name
	pn := routeField(r, "name")
	if pn == "" {
		writeError(w, errNoProviderName, http.StatusBadRequest)
		return
	}

	const (
		itemsPerPage = 10
		defaultPage = "1"
	)

	// if no page was passed so returns page 1
	pgStr := queryField(r, "page")
	if pgStr == "" {
		pgStr = defaultPage
	}

	page, err := strconv.Atoi(pgStr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	// retrieve all machine types
	mt, err := infra.DescribeMachineTypes()
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	// pages start at 1, can't be 0 or less.
	start := (page - 1) * itemsPerPage
	stop := start + itemsPerPage
	if start > len(mt) {
		writeError(w, errInvalidPageRange, http.StatusBadRequest)
		return
	}
	if stop > len(mt) {
		stop = len(mt)
	}

	writeResponse(w, mt[start:stop])
}

// HandleSSHKeys retrieves ssh keys from an infra provider
func (h *Handler) HandleSSHKeys(w http.ResponseWriter, r *http.Request) {
	pn := routeField(r, "name")
	if pn == "" {
		writeError(w, errNoProviderName, http.StatusBadRequest)
		return
	}

	// extract region
	region := queryField(r, "region")
	if region == "" {
		writeError(w, errReadQueryParam, http.StatusBadRequest)
		return
	}

	// retrieve ssh keys
	keys, err := infra.DescribeSSHKeys(region, h.DefaultConfig)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeResponse(w, keys)
}

func routeField(r *http.Request, field string) string {
	vars := mux.Vars(r)
	pn := vars[field]
	return pn
}

func queryField(r *http.Request, field string) (extracted string) {
	extracted = r.URL.Query().Get(field)
	return
}

type ErrResponder struct {
	Status  string `json:"status,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, err error, code int) {
	resp := ErrResponder{
		Status:  http.StatusText(code),
		Code:    code,
		Message: err.Error(),
	}
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeResponse(w http.ResponseWriter, body interface{}) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(body)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
	}
}
