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

	server := "localhost"
	port := "8080"

	// Read arguments from the command line
	// args := os.Args
	// if len(args) < 2 {
	// 	fmt.Println("Usage: go run server.go <server> <port>")
	// 	return
	// }

	// server := args[1]
	// port := args[2]

	// fmt.Println("Starting server on", server+":"+port)

	// fmt.Print("Enter the server address: ")
	// fmt.Scanln(&server)

	// fmt.Print("Enter the port number: ")
	// fmt.Scanln(&port)

	// Listen for incoming connections
	listener, err := net.Listen("tcp", server+":"+port)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Starting server on", server+":"+port)

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
