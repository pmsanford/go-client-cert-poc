package main

import (
	"log"
	"encoding/pem"
	"fmt"
	"os"
	"math/big"
	"crypto/rsa"
	"time"
	"crypto/x509/pkix"
	"crypto/x509"
	"crypto/rand"
)

func createTemplate(organization, country, state, city string, daysValid int) x509.Certificate {
	template := x509.Certificate {
		Subject: pkix.Name {
			Organization: []string{organization},
			Country: []string{country},
			Province: []string{state},
			Locality: []string{city},
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().Add(time.Duration(daysValid) * time.Hour * 24),
		BasicConstraintsValid: true,
		IsCA: true,
	}

	return template
}

func finishCert(cert x509.Certificate) ([]byte, *rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { return nil, nil, err }

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	cert.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil { return nil, nil, err }

	derBytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &priv.PublicKey, priv)
	return derBytes, priv, err
}

func createRootCert(organization, country, state, city, name, filename string, daysValid int) error {
	template := createTemplate(organization, country, state, city, daysValid)
	template.Subject.CommonName = name
	template.KeyUsage = x509.KeyUsageCertSign
	derBytes, priv, err := finishCert(template)
	if err != nil { return err }
	return writeCert(derBytes, priv, filename)
}

func writeCert(derBytes []byte, priv *rsa.PrivateKey, filename string) error {
	certOut, err := os.Create(fmt.Sprintf("%s.crt", filename))
	if err != nil { return err }
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(fmt.Sprintf("%s.key", filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil { return err }

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}


func testCA() {
	err := createRootCert("Test Org", "USA", "CA", "Mountain View", "Test Cert", "testcert", 365)
	if err != nil { log.Fatal(err) } else { log.Print("Success") }
}