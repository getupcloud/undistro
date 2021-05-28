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
package meta

const (
	LabelUndistroClusterName = "undistro.io/cluster-name"
	LabelUndistroClusterType = "undistro.io/cluster-type"
	LabelProviderType        = "undistro.io/provider-type"
	LabelUndistro            = "undistro.io"
	LabelUndistroMove        = "undistro.io/move"
	LabelUndistroMoved       = "undistro.io/moved"
	LabelUndistroInfra       = "node-role.undistro.io/infra"
	LabelK8sMaster           = "node-role.kubernetes.io/master"
	LabelK8sCP               = "node-role.kubernetes.io/control-plane"
)
