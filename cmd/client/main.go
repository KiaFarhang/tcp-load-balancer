package main

import (
	"io"
	"log"
	"net"
)

func main() {
	address, err := net.ResolveTCPAddr("tcp", ":4000")
	if err != nil {
		log.Fatalf("Error resolving TCP address: %s", err.Error())
	}

	connection, err := net.DialTCP("tcp", nil, address)

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
