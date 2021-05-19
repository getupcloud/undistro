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
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"
	"net/http"
)

type Metadata struct {
}


// RetrieveMetadata ...
func RetrieveMetadata(w http.ResponseWriter, r *http.Request) {
	// get provider name
	vars := mux.Vars(r)
	providerName, has := vars["name"]

	// validate provider name and type
	if !has {
		w.WriteHeader(http.StatusBadRequest)
		msg := struct {
			Status int
			ErrMessage string
		}{
			Status: http.StatusBadRequest,
			ErrMessage: "No provider name was found",
		}

		byt, err := json.Marshal(msg)
		if err != nil {
			klog.Errorln(err.Error())
			return
		}
		_, err = w.Write(byt)
		if err != nil {
			klog.Errorln(err.Error())
			return
		}
	}

	klog.Infoln("provider", providerName)

	var byt []byte
	read, err := r.Body.Read(byt)
	if err != nil {
		return
	}

	klog.Infoln("Read request body", read)
	//
}
