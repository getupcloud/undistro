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
package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/getupio-undistro/undistro/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Options is a struct for specifying configuration options for the render.Render object.
type Options struct {
	// Filesystem to load templates. Default clustertemplates Dir
	Filesystem fs.FS
	Root       string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
}

// Render is a service that provides functions for easily writing templates out to a writer.
type Render struct {
	// Customize Secure with an Options struct.
	opt       Options
	templates *template.Template
}

// New constructs a new Render instance with the supplied options.
func New(options ...Options) (*Render, error) {
	funcs := []template.FuncMap{sprig.TxtFuncMap()}
	var o Options
	if len(options) == 0 {
		o = Options{
			Funcs: funcs,
		}
	} else {
		o = options[0]
		o.Funcs = append(funcs, o.Funcs...)
	}

	r := Render{
		opt: o,
	}

	r.prepareOptions()
	err := r.compileTemplates()
	return &r, err
}

func (r *Render) prepareOptions() {
	if r.opt.Root == "" {
		r.opt.Root = "clustertemplates"
	}
	if r.opt.Filesystem == nil {
		r.opt.Filesystem = os.DirFS(".")
	}
}

func (r *Render) compileTemplates() error {
	r.templates = template.New(r.opt.Root)
	r.templates.Delims("{{", "}}")
	// Walk the supplied directory and compile any files that match our extension list.
	err := fs.WalkDir(r.opt.Filesystem, r.opt.Root, func(path string, info fs.DirEntry, err error) error { // nolint
		// Fix same-extension-dirs bug: some dir might be named to: "local.yaml".
		// These dirs should be excluded as they are not valid golang templates, but files under
		// them should be treat as normal.
		// If is a dir, return immediately (dir is not a valid golang template).
		if info == nil || info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		rel, err := filepath.Rel(r.opt.Root, path)
		if err != nil {
			return err
		}

		ext := ""
		extension := ".yaml"

		if strings.Contains(rel, ".") {
			ext = filepath.Ext(rel)
		}

		if ext == extension {
			buf, err := fs.ReadFile(r.opt.Filesystem, path)
			if err != nil {
				return err
			}

			name := (rel[0 : len(rel)-len(ext)])
			tmpl := r.templates.New(filepath.ToSlash(name))

			// Add our funcmaps.
			for _, funcs := range r.opt.Funcs {
				tmpl = tmpl.Funcs(funcs)
			}
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>", string(buf))
			_, err = tmpl.Parse(string(buf))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// TemplateLookup is a wrapper around template.Lookup and returns
// the template with the given name that is associated with t, or nil
// if there is no such template.
func (r *Render) TemplateLookup(t string) *template.Template {
	return r.templates.Lookup(t)
}

// Render is the generic function called by ClusterTemplate, and can be called by custom implementations.
func (r *Render) Render(w io.Writer, e Engine, data interface{}) error {
	return e.Render(w, data)
}

// YAML builds up the response from the specified template and bindings.
func (r *Render) YAML(w io.Writer, name string, binding interface{}) error {
	h := YAML{
		Name:      name,
		Templates: r.templates,
	}
	return r.Render(w, h, binding)
}

func GetObjs(fs fs.FS, dir, tplName string, vars map[string]interface{}) ([]unstructured.Unstructured, error) {
	tpl, err := New(Options{
		Root:       dir,
		Filesystem: fs,
	})
	if err != nil {
		return nil, err
	}
	buff := &bytes.Buffer{}
	err = tpl.YAML(buff, tplName, vars)
	if err != nil {
		return nil, err
	}
	objs, err := util.ToUnstructured(buff.Bytes())
	if err != nil {
		return nil, err
	}
	return objs, nil
}
