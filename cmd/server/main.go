package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
)

func main() {
	serverCert, err := tls.LoadX509KeyPair("certs/server/localhost.crt", "certs/server/localhost.key")
	if err != nil {
		log.Fatalf("Error loading server key pair: %s", err.Error())
	}

	caCert, err := ioutil.ReadFile("certs/ca/CA.pem")
	if err != nil {
		log.Fatalf("Error reading CA cert file: %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: caCertPool}

	listener, err := tls.Listen("tcp", "localhost:4000", tlsConfig)

	if err != nil {
		log.Fatalf("Error listening for TCP connections: %s", err.Error())
	}

	defer listener.Close()

	log.Println("Listening for TCP connections on port 4000...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting TCP connection: %s", err.Error())
		}

		go handleRequest(conn)
	}
}

type response struct {
	Message string `json:"message"`
}

func handleRequest(conn net.Conn) {
	log.Printf("Handling request from %s", conn.RemoteAddr())
	tlsConn, ok := conn.(*tls.Conn)
	defer conn.Close()
	if ok {

		// Server gets client cert after first i/o, so we explicitly call Handshake() to get the cert
		tlsConn.Handshake()
		logCertDetails(tlsConn.ConnectionState())
		response := &response{Message: "Yo girl"}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			conn.Write([]byte(err.Error()))
		} else {
			conn.Write(responseBytes)
		}
		//		conn.Write([]byte("Hi there"))
		log.Printf("Done handling request from %s", conn.RemoteAddr())
	} else {
		log.Print("Couldn't cast connection to TCP conn")

	}
}

func logCertDetails(state tls.ConnectionState) {
	log.Printf("Connection server name: %s", state.ServerName)
	for _, cert := range state.PeerCertificates {
		log.Printf("Email addresses: %s", cert.EmailAddresses)
	}
}
