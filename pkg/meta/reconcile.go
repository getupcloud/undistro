/*
Copyright 2020-2021 The UnDistro authors

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
	// ReconcileRequestAnnotation is the new ReconcileAtAnnotation, with a better name.
	ReconcileRequestAnnotation string = "reconcile.undistro.io/requestedAt"
	// finalizer undistro
	Finalizer string = "finalizer.undistro.io"
)

// ReconcileAnnotationValue returns a value for the reconciliation
// request annotations, which can be used to detect changes; and, a
// boolean indicating whether either annotation was set.
func ReconcileAnnotationValue(annotations map[string]string) (string, bool) {
	requestedAt, ok := annotations[ReconcileRequestAnnotation]
	return requestedAt, ok
}
