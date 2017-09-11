package main

import (
	"os"
	"encoding/json"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func register(client *http.Client, name string) *http.Client {
	resp, err := client.Get(fmt.Sprintf("https://localhost:8080/register?Name=%s", name))
	if err != nil {
		log.Printf("Error registering %s: %s", name, err)
	}
	defer resp.Body.Close()

	cont, err := ioutil.ReadAll(resp.Body)

	var newcert CertPair
	err = json.Unmarshal(cont, &newcert)

	if err != nil {
		log.Printf("Couldn't unmarshal cert pair for %s", name)
	}
	certfilename := fmt.Sprintf("%s.crt", name)
	keyfilename := fmt.Sprintf("%s.key", name)
	os.Remove(certfilename)
	os.Remove(keyfilename)

	certOut, err := os.Create(certfilename)
	if err != nil {
		log.Println("Couldn't open cert file.")
	}
	certOut.Write(newcert.Cert)
	certOut.Close()

	keyOut, err := os.Create(keyfilename)
	if err != nil {
		log.Println("Couldn't open key file.")
	}
	keyOut.Write(newcert.Key)
	keyOut.Close()

	clientcert, err := tls.LoadX509KeyPair(certfilename, keyfilename)
	if err != nil {
		log.Println("Couldn't load client cert:")
		log.Println(err)
		return nil
	}
	return createClient(&clientcert)
}

func doReq(client *http.Client, url string) {
	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	log.Printf("%s\n", string(contents))
	log.Println("")
}

func createClient(cert *tls.Certificate) *http.Client {
	// Load CA cert
	caCert, err := ioutil.ReadFile("root.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
	}
	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}

func runclient() {
	os.Remove("Paul.crt")
	os.Remove("Paul.key")
	os.Remove("Zac.crt")
	os.Remove("Zac.key")
	paulclient := createClient(nil)
	paulclient = register(paulclient, "Paul")
	doReq(paulclient, "https://localhost:8080/dothings")
	zacclient := createClient(nil)
	doReq(zacclient, "https://localhost:8080/dothings")
	zacclient = register(zacclient, "Zac")
	doReq(zacclient, "https://localhost:8080/dothings")
	doReq(paulclient, "https://localhost:8080/dothings")
	doReq(zacclient, "https://localhost:8080/dothings")
}
