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
package internalautotls

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
type Storage interface {
	Store(payload []byte) error
}

type FileStore struct {
}

func NewFileStorage() Storage {
	return &FileStore{}
}

func (fs *FileStore) Store(payload []byte) (err error) {
	err = os.WriteFile(path.Join("pki", "root.crt"), rootCertPEM, 0644)
	if err != nil {
		return errors.Errorf("unable to encode root crt to pem: %s", err.Error())
	}
	err = os.WriteFile(path.Join("pki", "root.key"), rootKeyPEM, 0644)
	if err != nil {
		return errors.Errorf("unable to encode root key to pem: %s", err.Error())
	}
	return nil
}

type SecretStore struct {
	client.Client
}

func NewSecretStore() Storage {
	return &SecretStore{}
}

func (s *SecretStore) Store(payload []byte) error {

	return nil
}
