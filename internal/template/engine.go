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
