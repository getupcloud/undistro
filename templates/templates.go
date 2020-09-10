/*
Copyright 2020 Getup Cloud. All rights reserved.
*/// Code generated for package templates by go-bindata DO NOT EDIT. (@generated)
// sources:
// templates/yaml/aws/cluster-template.yaml
package templates

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templatesYamlAwsClusterTemplateYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x56\x5b\x6b\xe3\x38\x18\x7d\xf7\xaf\x10\x26\xd0\x27\xbb\xcd\x2e\x94\xae\xdf\xd2\xd4\x6c\x4b\x12\x27\xd8\x6e\x4b\x59\x16\xa3\xda\x5f\x1c\x4d\x64\xcb\x48\x4a\xda\x12\xf2\xdf\x07\xc5\xb7\xdc\x9c\x5b\x67\x60\xfa\x54\x24\x7d\xb7\xe3\x73\xbe\x13\xc3\x30\x34\x9c\x91\x17\xe0\x82\xb0\xd4\x42\x21\x9d\x09\x09\xdc\xfc\x34\xa6\x77\xc2\x24\xec\x7a\xde\xc6\x34\x9b\xe0\xbf\xb5\x29\x49\x23\x0b\x75\xf3\x7b\x2d\x01\x89\x23\x2c\xb1\xa5\x21\x94\xe2\x04\x2c\xa4\xb7\x16\xdd\xfe\xb3\xe7\xdb\x6e\xe0\x74\x06\xf6\x52\xd7\x44\x06\xa1\xba\x2f\x72\x3a\x20\x3f\x18\x9f\xaa\x13\x84\x32\x16\x89\xfc\x3f\x84\x42\x12\xf1\x7b\xca\xc2\xa9\xb0\xd0\x7f\x7a\xfb\x9f\xbf\xcc\xf6\xed\x9d\x79\x63\xde\x5c\xb7\x6f\xf5\xff\x35\x84\x48\x3a\xe6\x58\x48\x3e\x0b\xe5\x8c\x83\x0b\xe3\x3c\x72\xbd\xef\xcd\x27\x66\xf3\x18\x2a\x30\x1f\xa5\xf3\xea\x95\xd3\xa8\xc3\xa6\x29\x10\x0a\x59\x2a\x39\xa3\x23\x8a\xd3\xba\x78\x9e\xa3\x37\x7b\x07\x1c\x25\xdd\xb5\x17\x3b\xad\x15\xe1\x99\xba\x3c\xd2\xd8\xfe\x1e\x8c\x22\x83\xb1\x4a\xa1\x6b\xdb\xdf\xec\xe4\xd9\x77\xe6\x3e\xe7\x2b\x72\x88\x57\xd5\xf4\xd6\xa2\xf3\xea\x05\xae\xfd\xef\xd3\xd0\x59\xe1\x23\xc4\xa4\x07\x5f\x4e\x19\xaf\x6e\x3d\xef\x31\xe8\xd9\x6f\x65\x0e\xd5\x71\x33\x60\x17\x80\x75\xbc\xf1\x6d\xd0\xea\x31\x32\x4a\x42\x2c\x2c\xd4\x5a\x74\x87\x8e\xef\x0e\xfb\xc1\xa8\xdf\x71\xec\x60\xd0\xe9\x3e\x3e\x39\x76\xd0\x1d\x3e\x3b\xfe\x72\x87\x76\x3e\x24\x19\xc5\x12\xac\x4d\x0a\x0d\x70\x38\x21\x69\x75\xfb\x3d\x62\x9e\x36\x0a\x42\xd3\x0a\xc5\x31\x89\xbd\x62\x32\xd5\x30\x91\xf9\xd9\x8c\x63\xa9\xca\x17\x0a\x4b\x59\x04\x2e\xc4\x44\xc8\xcd\xf3\xb2\xe2\xd5\x62\x81\x22\x61\x2a\x54\x03\x05\xab\x49\x59\x88\x69\x30\x61\x42\xaa\x07\x68\xb9\xbc\xaa\x22\x54\x6d\x0a\xd2\xfe\x94\x1c\x77\x78\x2c\xea\x5c\x4a\xe9\x6c\x16\x19\x19\x67\x73\x12\x01\xb7\x10\xfe\x10\x5a\x7e\xbe\x9a\x7a\x6f\x6f\x38\x23\x1e\xf0\x39\xf0\x3a\x11\x9c\x93\xbb\xd2\x27\x05\x3e\xc0\x29\x8e\x2f\xce\xf4\x83\x91\xf4\x0f\x84\x6f\x5e\x52\x49\x6f\x2d\x7a\xcf\xf7\xb6\xeb\xd8\xbe\xed\x05\x2f\xb6\xeb\xe5\x0a\xac\xd5\xb5\x87\x8f\x17\x71\xf1\x72\x79\xc9\x0d\x95\x88\x8a\x9a\x8a\x9c\x42\xe2\x34\x04\xff\x2b\xab\xf7\xc4\x7e\x05\xfa\x6f\xa3\x7c\xf3\xae\xe2\x70\xf2\x54\x84\x8e\x38\x1b\x13\xaa\xa2\x37\xaa\x97\x83\x18\x38\x23\x15\x7a\x06\xfe\x10\xa6\x20\xb1\x30\xf3\xe1\xca\x74\xa7\x2d\xab\x73\x2c\xb1\x40\xfc\x01\x32\xca\xbe\x12\x48\xe5\x29\xf0\x25\x91\x71\xb3\xeb\x90\x8d\x16\xb4\xbe\xb5\x5e\x87\x6e\xcf\x76\x77\xd7\x95\x00\x0a\xa1\x64\x05\xfb\x13\x2c\xc3\x49\x1f\xbf\x03\x15\x87\xbf\xcb\xd1\xda\xea\xef\x28\x07\xf3\x67\xef\x8c\x49\x25\x92\xac\x66\x75\xb8\x12\x54\x65\x9b\xeb\xb2\x69\x00\xa5\x7e\xb6\xfe\x15\xaa\xd4\x47\x56\x68\xa1\xb2\x2d\xb3\x19\x93\x78\x63\x43\x37\xfe\xa6\x38\xb9\xbb\x8b\x77\x3c\x3a\xe4\x1e\xdf\xb7\xf6\xed\x8c\x67\x73\xf1\x4c\x05\x3b\xc3\x87\x73\x84\xab\x56\xa9\xf8\xcd\x82\x3d\x85\x2a\x87\x08\xf2\x0b\x21\x6b\xb4\x94\x43\xa6\x72\x89\xad\x1c\x36\x96\xbd\xd6\xf2\x33\x00\x00\xff\xff\xeb\x8f\x7d\x52\xfb\x0b\x00\x00")

func templatesYamlAwsClusterTemplateYamlBytes() ([]byte, error) {
	return bindataRead(
		_templatesYamlAwsClusterTemplateYaml,
		"templates/yaml/aws/cluster-template.yaml",
	)
}

func templatesYamlAwsClusterTemplateYaml() (*asset, error) {
	bytes, err := templatesYamlAwsClusterTemplateYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/yaml/aws/cluster-template.yaml", size: 3067, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"templates/yaml/aws/cluster-template.yaml": templatesYamlAwsClusterTemplateYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"templates": &bintree{nil, map[string]*bintree{
		"yaml": &bintree{nil, map[string]*bintree{
			"aws": &bintree{nil, map[string]*bintree{
				"cluster-template.yaml": &bintree{templatesYamlAwsClusterTemplateYaml, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
