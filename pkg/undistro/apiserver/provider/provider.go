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
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider/infra/aws"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/rest"
)

var (
	errProviderNotSupported = errors.New("provider not supported yet")
	errInvalidProviderType  = errors.New("invalid provider type, supported are " +
		"['core', 'infra']")
)

// Provider wraps DescribeMetadata method
type Provider interface {
	DescribeMetadata() (interface{}, error)
}

// Handler holds rest config to access k8s endpoints
type Handler struct {
	DefaultConfig *rest.Config
}

func New(cfg *rest.Config) *Handler {
	return &Handler{
		DefaultConfig: cfg,
	}
}

const (
	ParamName     = "name"
	ParamType     = "type"
	ParamMeta     = "meta"
	ParamPage     = "page"
	ParamPageSize = "page_size"
	ParamRegion   = "region"
)

// HandleProviderMetadata retrieves Provider metadata by type
func (h *Handler) HandleProviderMetadata(w http.ResponseWriter, r *http.Request) {
	// extract provider type, infra provider as default
	providerType := queryProviderType(r)

	switch providerType {
	case string(configv1alpha1.InfraProviderType):

		// extract provider name
		providerName := queryField(r, ParamName)
		if isEmpty(providerName) {
			writeError(w, infra.ErrInvalidProviderName, http.StatusBadRequest)
			return
		}

		meta, err := extractMeta(r)
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}

		page, err := queryPage(r)
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}

		// extract provider region
		region := queryField(r, ParamRegion)

		// extract page size
		itemsPerPage, err := strconv.Atoi(queryField(r, ParamPageSize))
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}
		const defaultSize = 10
		if itemsPerPage <= 0 {
			itemsPerPage = defaultSize
		}

		infraProvider := infra.New(h.DefaultConfig, providerName, meta, region, page, itemsPerPage)
		resp, err := infraProvider.DescribeMetadata()
		if err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}

		writeResponse(w, resp)
	case string(configv1alpha1.CoreProviderType):
		// not supported yet
		writeError(w, errProviderNotSupported, http.StatusBadRequest)
	default:
		writeError(w, errInvalidProviderType, http.StatusBadRequest)
	}
}

func extractMeta(r *http.Request) (meta string, err error) {
	meta = queryField(r, ParamMeta)
	if isEmpty(meta) {
		err = aws.ErrNoProviderMeta
	}
	return
}

func queryField(r *http.Request, field string) (extracted string) {
	extracted = r.URL.Query().Get(field)
	return
}

func queryProviderType(r *http.Request) (providerType string) {
	providerType = queryField(r, ParamType)
	if isEmpty(providerType) {
		providerType = string(configv1alpha1.InfraProviderType)
	}
	return
}

func queryPage(r *http.Request) (page int, err error) {
	const defaultInitialPage = "1"
	pageSrt := queryField(r, ParamPage)
	switch {
	case !isEmpty(pageSrt):
		page, err = strconv.Atoi(pageSrt)
		if err != nil {
			return -1, err
		}
	default:
		page, err = strconv.Atoi(defaultInitialPage)
		if err != nil {
			return -1, err
		}
	}
	return
}

type errResponse struct {
	Status  string `json:"status,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, err error, code int) {
	resp := errResponse{
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

func isEmpty(s string) bool {
	return s == ""
}
