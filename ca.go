package main

import (
	"io"
	"io/ioutil"
	"net"
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

func finishCert(cert, rootCert *x509.Certificate, signingKey *rsa.PrivateKey) ([]byte, *rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil { return nil, nil, err }

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	cert.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil { return nil, nil, err }

	if signingKey == nil {
		signingKey = priv
	}

	if rootCert == nil {
		rootCert = cert
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, cert, rootCert, &priv.PublicKey, signingKey)
	return derBytes, priv, err
}

func createRootCert(organization, country, state, city, name, filename string, daysValid int) error {
	template := createTemplate(organization, country, state, city, daysValid)
	template.Subject.CommonName = name
	template.KeyUsage = x509.KeyUsageCertSign
	derBytes, priv, err := finishCert(&template, nil, nil)
	if err != nil { return err }
	return writeCert(derBytes, priv, filename)
}

func writeCert(derBytes []byte, priv *rsa.PrivateKey, filename string) error {
	certOut, err := os.Create(fmt.Sprintf("%s.crt", filename))
	if err != nil { return err }
	defer certOut.Close()
	keyOut, err := os.OpenFile(fmt.Sprintf("%s.key", filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil { return err }
	defer keyOut.Close()

	writeCertToBuffers(derBytes, priv, certOut, keyOut)

	return nil
}
func writeCertToBuffers(derBytes []byte, priv *rsa.PrivateKey, certOut, keyOut io.Writer) {
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

func genRoot() {
	os.Remove("*.crt")
	os.Remove("*.key")
	err := createRootCert("Test Org", "USA", "CA", "Mountain View", "Test Cert", "root", 365)
	if err != nil { log.Fatal(err) } else { log.Print("Created CA key") }
	err = createServerCert("Test Org", "USA", "CA", "Mountain View", "localhost", "127.0.0.1", "root", "server", 365)
	if err != nil { log.Fatal(err) } else { log.Print("Created Server key") }
}

func loadPem(filename string) (*pem.Block, error) {
	pemBytes, err := ioutil.ReadFile(filename)
	if err != nil { return nil, err }
	block, _ := pem.Decode(pemBytes)
	return block, nil 
}

func loadCertPrivate(filename string) (*x509.Certificate, *rsa.PrivateKey, error) {
	crt, err := loadCertPublic(filename)
	if err != nil { return nil, nil, err }
	keyBlock, err := loadPem(fmt.Sprintf("%s.key", filename))
	if err != nil { return nil, nil, err }
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil { return nil, nil, err }
	return crt, key, nil
}

func loadCertPublic(filename string) (*x509.Certificate, error) {
	crtBlock, err := loadPem(fmt.Sprintf("%s.crt", filename))
	if err != nil { return nil, err }
	crt, err := x509.ParseCertificate(crtBlock.Bytes)
	return crt, err
}

func createServerCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, filename string, daysValid int) error {
	return createAndWriteSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, filename, daysValid, x509.ExtKeyUsageServerAuth)
}

func createClientCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, filename string, daysValid int) error {
	return createAndWriteSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, filename, daysValid, x509.ExtKeyUsageClientAuth)
}

func getClientCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename string, daysValid int, certOut, keyOut io.Writer) (*big.Int, error) {
	derBytes, priv, serial, err := createSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, daysValid, x509.ExtKeyUsageClientAuth)
	if err != nil { return nil, err }
	writeCertToBuffers(derBytes, priv, certOut, keyOut)

	return serial, nil
}

func createAndWriteSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, filename string, daysValid int, keyUsage x509.ExtKeyUsage) error {
	derBytes, priv, _, err := createSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename, daysValid, keyUsage)
	if err != nil {
		return err
	}
	err = writeCert(derBytes, priv, filename)

	return err
}
func createSignedCert(organization, country, state, city, dnsName, ipAddr, rootCertFilename string, daysValid int, keyUsage x509.ExtKeyUsage) ([]byte, *rsa.PrivateKey, *big.Int, error) {
	template := createTemplate(organization, country, state, city, daysValid)
	template.Subject.CommonName = dnsName
	template.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	template.ExtKeyUsage = []x509.ExtKeyUsage { keyUsage }
	ipConverted := net.ParseIP(ipAddr)
	template.IPAddresses = append(template.IPAddresses, ipConverted)

	signingCert, signingKey, err := loadCertPrivate(rootCertFilename)

	if err != nil { return nil, nil, nil, err }

	 derBytes, priv, err := finishCert(&template, signingCert, signingKey)

	 return derBytes, priv, template.SerialNumber, err
}