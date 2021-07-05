package main

import (
	"crypto"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/smallstep/certificates/authority"
	"github.com/smallstep/certificates/authority/provisioner"
	"github.com/smallstep/truststore"
)

func main() {
	rp := caddy.NewReplacer()
	rp.Set("pki.ca.name", defaultCAName)
	rootCert, rootKey, err := generateRoot(rp.ReplaceAll(defaultRootCommonName, ""))
	if err != nil {
		panic(err)
	}
	rootCertPEM, err := pemEncodeCert(rootCert.Raw)
	if err != nil {
		panic(err)
	}
	rootKeyPEM, err := pemEncodePrivateKey(rootKey)
	if err != nil {
		panic(err)
	}
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
	csr, err := generateCSR(rootKey.(crypto.PrivateKey), []string{"localhost:8080", "127.0.0.1:8080"})
	if err != nil {
		panic(err)
	}
	certChain, err := auth.Sign(csr, provisioner.SignOptions{}, lft)
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "TLS")
	})
	http.ListenAndServeTLS(":8080", "./pki/certificate.crt", "./pki/root.key", nil)
}
