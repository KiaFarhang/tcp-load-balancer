package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

func main() {
	serverCert, err := tls.LoadX509KeyPair("certs/server/cert.pem", "certs/server/key.pem")
	if err != nil {
		log.Fatalf("Error loading server key pair: %s", err.Error())
	}

	clientAdminCert, err := ioutil.ReadFile("certs/client/admin-cert.pem")

	if err != nil {
		log.Fatalf("Error loading client admin cert: %s", err.Error())
	}

	clientUserCert, err := ioutil.ReadFile("certs/client/user-cert.pem")

	if err != nil {
		log.Fatalf("Error loading client user cert: %s", err.Error())
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientAdminCert)
	clientCertPool.AppendCertsFromPEM(clientUserCert)

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientCAs: clientCertPool, ClientAuth: tls.RequireAndVerifyClientCert}

	listener, err := tls.Listen("tcp", "localhost: 3333", tlsConfig)
	if err != nil {
		log.Fatalf("Error listening for TCP connections: %s", err.Error())
	}

	defer listener.Close()

	log.Println("Listening for TCP connections on port 3333...")
}
