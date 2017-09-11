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

func register(client *http.Client, name string) {
	resp, err := client.Get(fmt.Sprintf("https://localhost:8080/register?Name=%s", name))
	if err != nil {
		log.Printf("Error registering %s: %s", name, err)
	}
	defer resp.Body.Close()

	cont, err := ioutil.ReadAll(resp.Body)
	log.Printf("Content length: %d", len(cont))

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
	certOut.Write(newcert.cert)
	certOut.Close()

	keyOut, err := os.Create(certfilename)
	if err != nil {
		log.Println("Couldn't open key file.")
	}
	keyOut.Write(newcert.key)
	keyOut.Close()

	clientcert, err := tls.LoadX509KeyPair(certfilename, keyfilename)
	if err != nil {
		log.Println("Couldn't load client cert")
		return
	}
	transport := client.Transport.(*http.Transport)
	transport.TLSClientConfig.Certificates = []tls.Certificate{clientcert}
	transport.TLSClientConfig.BuildNameToCertificate()
	client.Transport = transport
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

func createClient() *http.Client {
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
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}

func runclient() {
	paulclient := createClient()
	//ericclient := createClient()
	register(paulclient, "Paul")
	/*
	doReq(paulclient, "https://localhost:8080/register?Name=Paul")
	doReq(ericclient, "https://localhost:8080/register?Name=Eric")
	doReq(paulclient, "https://localhost:8080/dothings")
	doReq(ericclient, "https://localhost:8080/dothings")
	*/
}
