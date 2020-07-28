/*
Copyright 2019 The Kubernetes Authors.

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

/*
Package cloudinit defines cloud init adapter for kind nodes and controlplane.

The Adapter supports a limited set of cloud init features, just what is necessary to test CPBPK;
additionally, for sake of simplicity, the adapter is designed to work on existing kind node images.
*/
package cloudinit
