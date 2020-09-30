package template

import (
	"bufio"
	"io"
	"text/template"

	"k8s.io/apimachinery/pkg/util/yaml"
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
	b, err := yaml.NewYAMLReader(bufio.NewReader(out)).Read()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}

	// Return the buffer to the pool.
	bufPool.Put(out)
	return nil
}
