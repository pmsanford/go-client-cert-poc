package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func doReq(client *http.Client, url string) {
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	contents, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s\n", string(contents))
	fmt.Println("")
}

func createClient(crtfile, keyfile string) *http.Client {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(crtfile, keyfile)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}

func main() {
	paulclient := createClient("paul.crt", "paul.key")
	ericclient := createClient("eric.crt", "eric.key")
	doReq(paulclient, "https://localhost:8080/register?Name=Paul")
	doReq(ericclient, "https://localhost:8080/register?Name=Eric")
	doReq(paulclient, "https://localhost:8080/dothings")
	doReq(ericclient, "https://localhost:8080/dothings")
}
