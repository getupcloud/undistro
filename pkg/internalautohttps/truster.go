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

package internalautohttps

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/getupio-undistro/undistro/pkg/undistro"
	"github.com/pkg/errors"
	"github.com/smallstep/truststore"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InstallLocalCert(ctx context.Context, c client.Client) (err error) {
	const caSecretName = "ca-secret"
	const caName = "ca.crt"
	objKey := client.ObjectKey{
		Namespace: undistro.Namespace,
		Name:      caSecretName,
	}
	secret := corev1.Secret{}
	err = c.Get(ctx, objKey, &secret)
	if err != nil {
		return errors.Errorf("unable to get CA secret %s: %v", caSecretName, err)
	}
	crtByt := secret.Data[caName]
	block, _ := pem.Decode(crtByt)
	rootCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Errorf("unable to parse certificate %s: %v", caName, err)
	}

	if !trusted(rootCert) {
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			return truststore.Install(rootCert,
				truststore.WithDebug(),
				truststore.WithFirefox(),
				truststore.WithJava(),
			)
		})
		if err != nil {
			return errors.Errorf("unable to install certificate %s: %v", caName, err)
		}
	}
	return
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}
