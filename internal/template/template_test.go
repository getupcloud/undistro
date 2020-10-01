package template

import (
	"bytes"
	"errors"
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
			render := New(Options{
				Directory: tc.directory,
				Funcs:     tc.funcs,
			})
			buff := bytes.Buffer{}
			err := render.YAML(&buff, tc.fileName, tc.values)
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(buff.String()).To(Equal(tc.out))
		})
	}
}

func TestFromAssets(t *testing.T) {
	g := NewWithT(t)
	render := New(Options{
		Asset: func(file string) ([]byte, error) {
			switch file {
			case "clustertemplates/test.yaml":
				return []byte("testassets: test"), nil
			default:
				return nil, errors.New("file not found: " + file)
			}
		},
		AssetNames: func() []string {
			return []string{"clustertemplates/test.yaml"}
		},
	})
	buff := bytes.Buffer{}
	err := render.YAML(&buff, "test", nil)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(buff.String()).To(Equal("testassets: test"))
}

func TestRace(t *testing.T) {
	g := NewWithT(t)
	render := New(Options{
		Directory: "testdata/basic",
	})
	done := make(chan struct{})
	req := func() {
		buff := bytes.Buffer{}
		err := render.YAML(&buff, "hello", "k8s")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(buff.String()).To(Equal("hello: test-k8s"))
		done <- struct{}{}
	}
	go req()
	go req()
	<-done
	<-done
}
