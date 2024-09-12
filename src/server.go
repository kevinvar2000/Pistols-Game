package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {

	// Buffer for incoming data
	buffer := make([]byte, 1024)

	for {
		// Read the incoming connection
		msg, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// Output the received message
		fmt.Println("Received message:", string(buffer[:msg]))

		// Send a response back to the client
		conn.Write([]byte("Message received"))
	}

}

func main() {

	// Listen for incoming connections
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port 8080...")

	// Accept connections in a loop
	for {

		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}

		// Handle connection in a separate goroutine for concurrency
		go handleConnection(conn)
	}

}
