package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	caCert, err := ioutil.ReadFile("certs/ca/CA.pem")
	if err != nil {
		log.Fatalf("Error reading CA cert file: %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	server := &http.Server{Addr: ":4000", Handler: http.HandlerFunc(handleRequest), TLSConfig: &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert, RootCAs: caCertPool}}

	log.Fatal(server.ListenAndServeTLS("certs/server/localhost.crt", "certs/server/localhost.key"))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	logCertDetails(req.TLS)
	fmt.Fprint(w, "Hello from HTTPS")
}

func logCertDetails(state *tls.ConnectionState) {
	log.Printf("Connection server name: %s", state.ServerName)
	for _, cert := range state.PeerCertificates {
		log.Printf("Email addresses: %s", cert.EmailAddresses)
	}
}
