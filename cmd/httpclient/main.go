package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

	transport := &http.Transport{TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{clientCert}, RootCAs: caCertPool}}

	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	response, err := client.Get("https://localhost:4000")

	if err != nil {
		log.Fatalf("Error calling server: %s", err.Error())
	}

	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Error reading response body: %s", err.Error())
	}

	fmt.Printf("Received response from server: %s\n", string(bodyBytes))

}
