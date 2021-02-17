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
	"os"
	"strings"
	"testing"
	"text/template"

	. "github.com/onsi/gomega"
)

func upperTest() string {
	return strings.ToUpper("test")
}

func TestRender(t *testing.T) {
	os.Setenv("TEST_ENV", "env")
	defer os.Clearenv()
	testCases := []struct {
		name      string
		directory string
		fileName  string
		wantErr   bool
		out       string
		values    interface{}
		funcs     []template.FuncMap
	}{
		{
			name:      "file not exists",
			directory: "testdata/basic",
			fileName:  "nope",
			wantErr:   true,
		},
		{
			name:      "directory not exists",
			directory: "../testdata/basic",
			fileName:  "hello",
			wantErr:   true,
		},
		{
			name:      "valid",
			directory: "testdata/basic",
			fileName:  "hello",
			wantErr:   false,
			out:       "hello: test-k8s",
			values:    "k8s",
		},
		{
			name:      "valid import",
			directory: "testdata/basic",
			fileName:  "import",
			wantErr:   false,
			out:       "hello: test-k8s\nadmin: test-k8s",
			values:    "k8s",
		},
		{
			name:      "custom func",
			directory: "testdata/funcs",
			fileName:  "funcs",
			wantErr:   false,
			out:       "testfunc: test-TEST\ntestdefault: test-env",
			funcs:     []template.FuncMap{{"func": upperTest}},
		},
		{
			name:      "default func",
			directory: "testdata/basic",
			fileName:  "funcs-default",
			wantErr:   false,
			out:       "testdefault: test-env",
		},
		{
			name:      "nested",
			directory: "testdata/basic",
			fileName:  "admin/admin",
			wantErr:   false,
			out:       "admin: test-k8s",
			values:    "k8s",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			render, err := New(Options{
				Root:       tc.directory,
				Funcs:      tc.funcs,
				Filesystem: os.DirFS(tc.directory),
			})
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
			}
			buff := bytes.Buffer{}
			err = render.YAML(&buff, tc.fileName, tc.values)
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(buff.String()).To(Equal(tc.out))
		})
	}
}

func TestRace(t *testing.T) {
	g := NewWithT(t)
	render, err := New(Options{
		Root:       "testdata/basic",
		Filesystem: os.DirFS("testdata/basic"),
	})
	g.Expect(err).ToNot(HaveOccurred())
	done := make(chan struct{})
	req := func() {
		buff := bytes.Buffer{}
		err = render.YAML(&buff, "hello", "k8s")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(buff.String()).To(Equal("hello: test-k8s"))
		done <- struct{}{}
	}
	go req()
	go req()
	<-done
	<-done
}
