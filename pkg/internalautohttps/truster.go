package internalautohttps

import (
	"context"
	"crypto/x509"

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
	rootCert, err := x509.ParseCertificate(crtByt)
	if err != nil {
		return errors.Errorf("unable to parse certificate %s: %v", caName, err)
	}

	if !trusted(rootCert) {
		truststore.Install(rootCert,
			truststore.WithDebug(),
			truststore.WithFirefox(),
			truststore.WithJava(),
		)
	}
	return
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}
