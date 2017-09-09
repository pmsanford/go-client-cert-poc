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

func OpenDb() (*sql.DB) {
	db, err := sql.Open("sqlite3", "./reg.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	user := ValidateCert(req)
	if user != nil {
		log.Printf("Found user %s", user.name)
		fmt.Fprintf(w, "Certificate with serial %s is already registered to %s", user.serial, user.name)
	} else {
		log.Println("No row found")
		db := OpenDb()
		defer db.Close()
		user = ParseCert(req)
		_, err := db.Exec("insert into registrations (serial, name) values (?, ?)", user.serial, user.name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, "Registered %s with serial # %s", user.name, user.serial)
	}
}

func ParseCert(req *http.Request) *User {
	serial := req.TLS.PeerCertificates[0].SerialNumber.String()
	name := req.URL.Query()["Name"][0]
	return &User { serial: serial, name: name }
}

type User struct {
	serial string
	name string
}

func ValidateCert(req *http.Request) *User {
	db := OpenDb()
	defer db.Close()
	serial := req.TLS.PeerCertificates[0].SerialNumber.String()
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
