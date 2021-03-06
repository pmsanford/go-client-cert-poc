package main

import (
	"strings"
	"encoding/json"
	"bytes"
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

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	user := ParseCert(req)
	if valid, regname := ValidateCert(*user); valid {
		log.Printf("Found user %s", regname)
		fmt.Fprintf(w, "Certificate with serial %s is already registered to %s", user.serial, regname)
	} else {
		log.Println("No row found")
		var certOut bytes.Buffer
		var keyOut bytes.Buffer
		ipstr := req.RemoteAddr
		if ipstr[0:1] == "[" {
			ipstr = ipstr[1:]
			ipstr = strings.Split(ipstr, "]")[0]
		} else {
			ipstr = strings.Split(ipstr, ":")[0]
		}
		serial, err := getClientCert("Test Org", "USA", "CA", "Mountain View", user.name, ipstr, "root", 365, &certOut, &keyOut)
		logErr(err)
		certBuf, err := ioutil.ReadAll(&certOut)
		logErr(err)
		keyBuf, err := ioutil.ReadAll(&keyOut)
		logErr(err)
		pair := CertPair {
			Cert: certBuf,
			Key: keyBuf,
		}
		db := OpenDb()
		defer db.Close()
		_, err = db.Exec("insert into registrations (serial, name) values (?, ?)", serial.String(), user.name)
		if err != nil {
			log.Fatal(err)
		}
		jsonBytes, err := json.Marshal(pair)
		logErr(err)
		w.Write(jsonBytes)
	}
}

func ParseCert(req *http.Request) *User {
	serial := ""
	if len(req.TLS.PeerCertificates) > 0 { 
		serial = req.TLS.PeerCertificates[0].SerialNumber.String()
	}
	name := ""
	if names, ok := req.URL.Query()["Name"]; ok {
		name = names[0]
	}
	return &User { serial: serial, name: name }
}

type User struct {
	serial string
	name string
}

func ValidateCert(requsr User) (bool, string) {
	db := OpenDb()
	defer db.Close()
	serial := requsr.serial
	rows, err := db.Query("select name from registrations where serial = ?", serial)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		log.Println("Found a row")
		var exname string
		rows.Scan(&exname)
		return true, exname
	}
	return false, ""
}

func DoOp(w http.ResponseWriter, req *http.Request) {
	user := ParseCert(req)
	if ok, name := ValidateCert(*user); ok {
		fmt.Fprintf(w, "Doing things with %s", name)
	} else {
		fmt.Fprintf(w, "Please register; no user found for cert with serial %s", user.serial)
	}
}

func runserver() {
	os.Remove("reg.db")
	db, err := sql.Open("sqlite3", "./reg.db")
	if err != nil {
		fmt.Println(err)
	}
	os.Remove("root.crt")
	os.Remove("root.key")
	os.Remove("server.crt")
	os.Remove("server.key")
	genRoot()

	query := "create table registrations (serial text not null primary key, name text); delete from registrations;"
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
	}
	db.Close()
	http.HandleFunc("/register", HelloServer)
	http.HandleFunc("/dothings", DoOp)

	caCertRoot, err := ioutil.ReadFile("root.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertRoot); !ok {
		log.Fatal("Couldn't append certs")
	}

	tlsConfig := &tls.Config{
		ClientCAs: caCertPool,
		ClientAuth: tls.RequestClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:      ":8080",
		TLSConfig: tlsConfig,
	}

	log.Println("Starting server")

	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key")) //private cert
}
