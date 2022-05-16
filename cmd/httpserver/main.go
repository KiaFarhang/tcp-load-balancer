package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func main() {
	server := &http.Server{Addr: ":4000", Handler: http.HandlerFunc(handleRequest), TLSConfig: &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAnyClientCert}}

	log.Fatal(server.ListenAndServeTLS("certs/server/localhost.crt", "certs/server/localhost.key"))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	logCertDetails(req.TLS)
	fmt.Fprint(w, "Hello from HTTPS")
}

func logCertDetails(state *tls.ConnectionState) {
	log.Printf("Connection server name: %s", state.ServerName)
	for _, cert := range state.PeerCertificates {
		extensions := cert.Extensions
		log.Printf("First extension key: %s, value: %s", extensions[0].Id, extensions[0].Value)
		log.Printf("Email addresses: %s", cert.EmailAddresses)
	}
}
