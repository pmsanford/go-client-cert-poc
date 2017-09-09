package main

import (
	"database/sql"
	"crypto/tls"
	"crypto/x509"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	db, _ := sql.Open("sqlite3", "./reg.db")
	defer db.Close()
	serial := req.TLS.PeerCertificates[0].SerialNumber.String()
	name := req.URL.Query()["Name"][0]
	rows, err := db.Query("select name from registrations where serial = ?", serial)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		log.Println("Found a row")
		var exname string
		rows.Scan(&exname)
		fmt.Fprintf(w, "Certificate with serial %s is already registered to %s", serial, exname)
	} else {
		log.Println("No row found")
		_, err := db.Exec("insert into registrations (serial, name) values (?, ?)", serial, name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "Registered %s with serial # %s", name, serial)
	}
}

type User struct {
	serial string
	name string
}

func ValidateCert(req *http.Request) *User {
	db, _ := sql.Open("sqlite3", "./reg.db")
	defer db.Close()
	serial := req.TLS.PeerCertificates[0].SerialNumber.String()
	name := req.URL.Query()["Name"][0]
	rows, err := db.Query("select name from registrations where serial = ?", serial)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		log.Println("Found a row")
		var exname string
		rows.Scan(&exname)
		return &User { serial: serial, name: exname }
	}
	return nil
}

func DoOp(w http.ResponseWriter, req *http.Request) {

}

func main() {
	os.Remove("./reg.db")
	db, err := sql.Open("sqlite3", "./reg.db")
	if err != nil {
		fmt.Println(err)
	}

	query := "create table registrations (serial text not null primary key, name text); delete from registrations;"
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	}
	db.Close()
	http.HandleFunc("/register", HelloServer)

	caCert, err := ioutil.ReadFile("client.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		// NoClientCert
		// RequestClientCert
		// RequireAnyClientCert
		// VerifyClientCertIfGiven
		// RequireAndVerifyClientCert
		ClientAuth: tls.RequestClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      ":8080",
		TLSConfig: tlsConfig,
	}

	server.ListenAndServeTLS("server.crt", "server.key") //private cert
}
