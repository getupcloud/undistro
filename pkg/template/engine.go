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
package template

import (
	"io"
	"text/template"

	"sigs.k8s.io/yaml"
)

// Engine is the generic interface for all responses.
type Engine interface {
	Render(io.Writer, interface{}) error
}

// YAML built-in renderer.
type YAML struct {
	Name      string
	Templates *template.Template
}

// Render a HTML response.
func (y YAML) Render(w io.Writer, binding interface{}) error {
	// Retrieve a buffer from the pool to write to.
	out := bufPool.Get()
	err := y.Templates.ExecuteTemplate(out, y.Name, binding)
	if err != nil {
		return err
	}
	_, err = yaml.YAMLToJSON(out.Bytes())
	if err != nil {
		return err
	}
	_, err = out.WriteTo(w)
	if err != nil {
		return err
	}

	// Return the buffer to the pool.
	bufPool.Put(out)
	return nil
}
