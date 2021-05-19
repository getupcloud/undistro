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
package apiserver

import (
	"context"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/provider"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getupio-undistro/undistro/pkg/fs"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/health"
	"github.com/getupio-undistro/undistro/pkg/undistro/apiserver/proxy"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Server struct {
	genericclioptions.IOStreams
	*http.Server
	K8sCfg        *rest.Config
	HealthHandler health.Handler
}

func NewServer(cfg *rest.Config, in io.Reader, out, errOut io.Writer, healthChecks ...health.Checker) *Server {
	streams := genericclioptions.IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}
	router := mux.NewRouter()
	apiServer := &Server{
		IOStreams: streams,
		Server: &http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		K8sCfg: cfg,
	}
	for _, c := range healthChecks {
		apiServer.HealthHandler.Add(c)
	}
	apiServer.routes(router)
	env := os.Getenv("UNDISTRO_ENV")
	apiServer.Handler = router
	if env == "dev" {
		apiServer.Handler = handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPatch,
				http.MethodPut,
				http.MethodOptions,
				http.MethodHead,
				http.MethodDelete,
				http.MethodConnect,
			}),
		)(router)
	}
	return apiServer
}

func (s *Server) routes(router *mux.Router) {
	router.Handle("/healthz/readiness", &s.HealthHandler)
	router.HandleFunc("/healthz/liveness", health.HandleLive)
	router.HandleFunc("/provider/{name}/metadata", provider.RetrieveMetadata)
	router.PathPrefix("/uapi/v1/namespaces/{namespace}/clusters/{cluster}/proxy/").Handler(proxy.NewHandler(s.K8sCfg))
	router.PathPrefix("/").Handler(fs.ReactHandler("", "frontend"))
}

func (s *Server) GracefullyStart(ctx context.Context, addr string) error {
	s.Addr = addr
	cerr := make(chan error, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func(ctx context.Context) {
		klog.Infof("listen on %s", addr)
		cerr <- s.ListenAndServe()
	}(ctx)
	select {
	case <-sigCh:
		return s.Shutdown(ctx)
	case err := <-cerr:
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
