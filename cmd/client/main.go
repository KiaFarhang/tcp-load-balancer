package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
)

func main() {
	clientCert, err := tls.LoadX509KeyPair("certs/client/admin.crt", "certs/client/client.key")

	if err != nil {
		log.Fatalf("Error loading client cert: %s", err.Error())
	}

	caCert, err := ioutil.ReadFile("certs/ca/CA.pem")
	if err != nil {
		log.Fatalf("Error reading CA cert file: %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{RootCAs: caCertPool, Certificates: []tls.Certificate{clientCert}}

	connection, err := tls.Dial("tcp", "localhost:4000", tlsConfig)

	if err != nil {
		log.Fatalf("Error dialing TCP: %s", err.Error())
	}

	defer connection.Close()

	responseBytes, err := io.ReadAll(connection)

	if err != nil {
		log.Fatalf("Error reading response: %s", err.Error())
	}

	log.Printf("Response from server: %s", string(responseBytes))

}
