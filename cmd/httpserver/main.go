package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func main() {
	server := &http.Server{Addr: ":4000", Handler: http.HandlerFunc(handleRequest), TLSConfig: &tls.Config{MinVersion: tls.VersionTLS13}}

	log.Fatal(server.ListenAndServeTLS("certs/server/localhost.crt", "certs/server/localhost.key"))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello from HTTPS")
}
