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
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
	"github.com/pkg/errors"
	"github.com/smallstep/certificates/authority"
	"github.com/smallstep/certificates/authority/provisioner"
	"github.com/smallstep/cli/crypto/x509util"
	"github.com/smallstep/truststore"
	"golang.org/x/net/idna"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	rootCert                      = "root.cert"
	rootKey                       = "root.key"
	defaultCAName                 = "UnDistro Local Authority"
	defaultRootCommonName         = "{pki.ca.name} - {time.now.year} ECC Root"
	defaultIntermediateCommonName = "{pki.ca.name} - ECC Intermediate"
	defaultRootLifetime           = 24 * time.Hour * 30 * 12 * 10
	defaultIntermediateLifetime   = 24 * time.Hour * 7
	defaultInternalCertLifetime   = 12 * time.Hour
)

type UndistroIssuer interface {
	Issue([]string) error
}

type InternalIssuer struct {
	CertificateAuthority string
	RootLifetime         time.Duration
	IntermediateLifetime time.Duration
	InternalCertLifetime time.Duration
	genericclioptions.IOStreams
}

func New() *InternalIssuer {
	return &InternalIssuer{
		CertificateAuthority: defaultCAName,
		RootLifetime:         defaultRootLifetime,
		IntermediateLifetime: defaultIntermediateLifetime,
		InternalCertLifetime: defaultInternalCertLifetime,
	}
}

// sans
func (at *InternalIssuer) Issue([]string) error {
	rp := caddy.NewReplacer()
	rp.Set("pki.ca.name", defaultCAName)
	rootCert, rootKey, err := generateRoot(rp.ReplaceAll(defaultRootCommonName, ""))
	if err != nil {
		return errors.Errorf("unable to generate root certs: %s", err.Error())
	}
	rootCertPEM, err := pemEncodeCert(rootCert.Raw)
	if err != nil {
		return errors.Errorf("unable to encode root cert to pem: %s", err.Error())
	}
	rootKeyPEM, err := pemEncodePrivateKey(rootKey)
	if err != nil {
		return errors.Errorf("unable to encode root key to pem: %s", err.Error())
	}

	// create a secret in k8s
	err = os.WriteFile(path.Join("pki", "root.crt"), rootCertPEM, 0644)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path.Join("pki", "root.key"), rootKeyPEM, 0644)
	if err != nil {
		panic(err)
	}

	if !trusted(rootCert) {
		truststore.Install(rootCert,
			truststore.WithDebug(),
			truststore.WithFirefox(),
			truststore.WithJava(),
		)
	}
	opts := []authority.Option{
		authority.WithX509Signer(rootCert, rootKey.(crypto.Signer)),
		authority.WithX509RootCerts(rootCert),
	}
	auth, err := authority.NewEmbedded(opts...)
	if err != nil {
		panic(err)
	}
	lft := defaultInternalCertLifetime
	// ensure issued certificate does not expire later than its issuer
	if time.Now().Add(lft).After(rootCert.NotAfter) {
		lft = time.Until(rootCert.NotAfter)
	}

	csr, err := generateCSR(rootKey.(crypto.PrivateKey), []string{"localhost", "127.0.0.1", "undistro.local"})
	if err != nil {
		panic(err)
	}
	certChain, err := auth.Sign(csr, provisioner.SignOptions{}, customCertLifetime(lft))
	if err != nil {
		panic(err)
	}

	f, err := os.Create(path.Join("pki", "certificate.crt"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, cert := range certChain {
		err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		if err != nil {
			panic(err)
		}
	}
}

type customCertLifetime time.Duration

func (d customCertLifetime) Modify(cert *x509.Certificate, _ provisioner.SignOptions) error {
	cert.NotBefore = time.Now()
	cert.NotAfter = cert.NotBefore.Add(time.Duration(d))
	return nil
}

func (at *InternalIssuer) Renew() {

}

func pemEncodeCert(der []byte) ([]byte, error) {
	return pemEncode("CERTIFICATE", der)
}

// pemEncodePrivateKey marshals a EC or RSA private key into a PEM-encoded array of bytes.
// TODO: this is the same thing as in certmagic. Should we reuse that code somehow? It's unexported.

func pemEncodePrivateKey(key crypto.PrivateKey) ([]byte, error) {
	var pemType string
	var keyBytes []byte
	switch key := key.(type) {
	case *ecdsa.PrivateKey:
		var err error
		pemType = "EC"
		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
	case *rsa.PrivateKey:
		pemType = "RSA"
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
	case *ed25519.PrivateKey:
		var err error
		pemType = "ED25519"
		keyBytes, err = x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}
	return pemEncode(pemType+" PRIVATE KEY", keyBytes)
}

func pemEncode(blockType string, b []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{Type: blockType, Bytes: b})
	return buf.Bytes(), err
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}

func generateRoot(commonName string) (rootCrt *x509.Certificate, privateKey interface{}, err error) {
	rootProfile, err := x509util.NewRootProfile(commonName)
	if err != nil {
		return
	}
	rootProfile.Subject().NotAfter = time.Now().Add(defaultRootLifetime) // TODO: make configurable
	return newCert(rootProfile)
}

func generateIntermediate(commonName string, rootCrt *x509.Certificate, rootKey interface{}) (cert *x509.Certificate, privateKey interface{}, err error) {
	interProfile, err := x509util.NewIntermediateProfile(commonName, rootCrt, rootKey)
	if err != nil {
		return
	}
	interProfile.Subject().NotAfter = time.Now().Add(defaultIntermediateLifetime) // TODO: make configurable
	return newCert(interProfile)
}

func newCert(profile x509util.Profile) (cert *x509.Certificate, privateKey interface{}, err error) {
	certBytes, err := profile.CreateCertificate()
	if err != nil {
		return
	}
	privateKey = profile.SubjectPrivateKey()
	cert, err = x509.ParseCertificate(certBytes)
	return
}

func generateCSR(privateKey crypto.PrivateKey, sans []string) (*x509.CertificateRequest, error) {
	csrTemplate := new(x509.CertificateRequest)

	for _, name := range sans {
		if ip := net.ParseIP(name); ip != nil {
			csrTemplate.IPAddresses = append(csrTemplate.IPAddresses, ip)
		} else if strings.Contains(name, "@") {
			csrTemplate.EmailAddresses = append(csrTemplate.EmailAddresses, name)
		} else if u, err := url.Parse(name); err == nil && strings.Contains(name, "/") {
			csrTemplate.URIs = append(csrTemplate.URIs, u)
		} else {
			// convert IDNs to ASCII according to RFC 5280 section 7
			normalizedName, err := idna.ToASCII(name)
			if err != nil {
				return nil, fmt.Errorf("converting identifier '%s' to ASCII: %v", name, err)
			}
			csrTemplate.DNSNames = append(csrTemplate.DNSNames, normalizedName)
		}
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificateRequest(csrDER)
}

var (
	_ certmagic.Issuer = (*InternalIssuer)(nil)
)
