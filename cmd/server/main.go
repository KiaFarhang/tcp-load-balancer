package main

import (
	"log"
	"net"
)

func main() {
	// serverCert, err := tls.LoadX509KeyPair("certs/server/localhost.crt", "certs/server/localhost.key")
	// if err != nil {
	// 	log.Fatalf("Error loading server key pair: %s", err.Error())
	// }

	// clientCertPool := x509.NewCertPool()

	// tlsConfig := &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientCAs: clientCertPool, ClientAuth: tls.RequireAndVerifyClientCert}

	listener, err := net.Listen("tcp", "localhost:4000")

	//listener, err := tls.Listen("tcp", "localhost: 3333", tlsConfig)
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

func handleRequest(conn net.Conn) {
	log.Printf("Handling request from %s", conn.RemoteAddr())
	//io.Copy(conn, conn)
	conn.Write([]byte("Hi there"))
	conn.Close()
	log.Printf("Done handling request from %s", conn.RemoteAddr())
}
