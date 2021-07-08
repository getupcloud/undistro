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
	"context"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"

	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//
type Storage interface {
	Store(chain []*x509.Certificate, rootCert, rootKey []byte) error
}

type FileStore struct{}

func NewFileStorage() Storage {
	return &FileStore{}
}

func (fs *FileStore) Store(chain []*x509.Certificate, rootCert, rootKey []byte) (err error) {
	err = os.WriteFile(path.Join("pki", "root.crt"), rootCert, 0644)
	if err != nil {
		return errors.Errorf("unable to encode root crt to pem: %s", err.Error())
	}
	err = os.WriteFile(path.Join("pki", "root.key"), rootKey, 0644)
	if err != nil {
		return errors.Errorf("unable to encode root key to pem: %s", err.Error())
	}

	//
	f, err := os.Create(path.Join("pki", "certificate.crt"))
	if err != nil {
		return errors.Errorf("unable to generate the Certicate Signing Request: %s", err.Error())
	}
	defer f.Close()
	for _, cert := range chain {
		err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			return errors.Errorf("unable to encode the certificate: %s", err.Error())
		}
	}
	return nil
}

type SecretStore struct {
	client.Client
}

func NewSecretStore() Storage {
	return &SecretStore{}
}

func (s *SecretStore) Store(chain []*x509.Certificate, rootCert, rootKey []byte) error {
	const secretName = ""

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
		Data: map[string][]byte{
			"ca.crt":  string(rootCert),
			"tls.crt": string(rootKey),
			"tls.key": "",
		},
	}
	_, err = util.CreateOrUpdate(context.Background(), s.Client, &secret)
	if err != nil {
		return err
	}
	return nil
}
