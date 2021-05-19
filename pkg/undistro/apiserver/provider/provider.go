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
	"encoding/json"
	"errors"
	"github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/gorilla/mux"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type Metadata struct {
}

type ErrResponse struct {
	Status  int `json:"status"`
	Message string `json:"message"`
}

var (
	NoProviderName = errors.New("no provider name was found")
	ReadBody = errors.New("error while parsing body")
	InvalidProviderName = errors.New("invalid provider, maybe unsupported")
)

// RetrieveMetadata ...
func RetrieveMetadata(w http.ResponseWriter, r *http.Request) {
	providerName := routeField("name", w, r)
	if providerName == "" {
		return
	}

	if !isValidInfraProvider(providerName) {
		http.Error(w, InvalidProviderName.Error(), http.StatusBadRequest)
		return
	}

	// get provider type
	var body struct{
		Type string
	}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			klog.Errorln(err.Error())
		}
	}(r.Body)

	if err := d.Decode(&body); err != nil {
		http.Error(w, ReadBody.Error(), http.StatusInternalServerError)
		return
	}

	klog.Infoln("body type", body.Type)

	//generate metadata info about the provider
}

func routeField(routeField string, w http.ResponseWriter, r *http.Request) (pn string) {
	vars := mux.Vars(r)
	pn, has := vars[routeField]
	if !has {
		http.Error(w, NoProviderName.Error(), http.StatusBadRequest)
	}
	return strings.TrimSpace(pn)
}

func isValidInfraProvider(name string) bool {
	return name == v1alpha1.Aws.String()
}
