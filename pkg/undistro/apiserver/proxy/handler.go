/*
Copyright 2020 The UnDistro authors

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
package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/gorilla/mux"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultNamespace = "undistro-system"
	defaultCluster   = "management"
)

type Handler struct {
	DefaultConfig *rest.Config
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cluster := vars["cluster"]
	namespace := vars["namespace"]
	mgmtClient, err := client.New(h.DefaultConfig, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		klog.Error(err)
		h.handleErrorCode(w, http.StatusInternalServerError)
		return
	}
	cfg := h.DefaultConfig
	if namespace != defaultNamespace || cluster != defaultCluster {
		cfg, err = kube.NewClusterConfig(r.Context(), mgmtClient, cluster, namespace)
		if err != nil {
			klog.Error(err)
			h.handleErrorCode(w, http.StatusBadRequest)
			return
		}
	}
	proxyPrefix := fmt.Sprintf("/uapi/v1/namespaces/%s/clusters/%s/proxy", namespace, cluster)
	proxy, err := kube.NewProxyHandler(proxyPrefix, cfg, time.Minute*30)
	if err != nil {
		klog.Error(err)
		h.handleErrorCode(w, http.StatusBadRequest)
		return
	}
	proxy.ServeHTTP(w, r)
}

func NewHandler(cfg *rest.Config) *Handler {
	return &Handler{
		DefaultConfig: cfg,
	}
}

func (h *Handler) handleErrorCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Write([]byte(http.StatusText(code)))
}
