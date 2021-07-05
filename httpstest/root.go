package main

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
	"strings"
	"time"

	"github.com/smallstep/cli/crypto/x509util"
	"golang.org/x/net/idna"
)

const (
	defaultCAName                 = "UnDistro Local Authority"
	defaultRootCommonName         = "{pki.ca.name} - {time.now.year} ECC Root"
	defaultIntermediateCommonName = "{pki.ca.name} - ECC Intermediate"

	defaultRootLifetime         = 24 * time.Hour * 30 * 12 * 10
	defaultIntermediateLifetime = 24 * time.Hour * 7
	defaultInternalCertLifetime = 12 * time.Hour
)

func pemDecodeSingleCert(pemDER []byte) (*x509.Certificate, error) {
	pemBlock, remaining := pem.Decode(pemDER)
	if pemBlock == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	if len(remaining) > 0 {
		return nil, fmt.Errorf("input contained more than a single PEM block")
	}
	if pemBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("expected PEM block type to be CERTIFICATE, but got '%s'", pemBlock.Type)
	}
	return x509.ParseCertificate(pemBlock.Bytes)
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

// pemDecodePrivateKey loads a PEM-encoded ECC/RSA private key from an array of bytes.
// Borrowed from Go standard library, to handle various private key and PEM block types.
// https://github.com/golang/go/blob/693748e9fa385f1e2c3b91ca9acbb6c0ad2d133d/src/crypto/tls/tls.go#L291-L308
// https://github.com/golang/go/blob/693748e9fa385f1e2c3b91ca9acbb6c0ad2d133d/src/crypto/tls/tls.go#L238)
// TODO: this is the same thing as in certmagic. Should we reuse that code somehow? It's unexported.
func pemDecodePrivateKey(keyPEMBytes []byte) (crypto.PrivateKey, error) {
	keyBlockDER, _ := pem.Decode(keyPEMBytes)

	if keyBlockDER.Type != "PRIVATE KEY" && !strings.HasSuffix(keyBlockDER.Type, " PRIVATE KEY") {
		return nil, fmt.Errorf("unknown PEM header %q", keyBlockDER.Type)
	}

	if key, err := x509.ParsePKCS1PrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(keyBlockDER.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping: %T", key)
		}
	}

	if key, err := x509.ParseECPrivateKey(keyBlockDER.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("unknown private key type")
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
