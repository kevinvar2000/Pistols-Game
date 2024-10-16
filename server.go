package main

import (
	"fmt"
	"net"
)

func (s *Server) addClient(client Client) {
	s.clients[client.name] = client
}

func (s *Server) removeClient(client Client) {
	delete(s.clients, client.name)
}

func (s *Server) broadcast(message string) {
	for _, client := range s.clients {
		client.conn.Write([]byte(message))
	}
}

func (s *Server) handleConnection(conn net.Conn) {

	// Buffer for incoming data
	buffer := make([]byte, 1024)

	// Read the client
	msg, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	// Create a new client
	client := Client{conn, string(buffer[:msg])}

	// Add the client to the server
	s.addClient(client)

	// Broadcast the new client
	s.broadcast(client.name + " has joined the chat")

	for {
		// Read the incoming connection
		msg, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// Output the received message
		fmt.Println("Received message:", string(buffer[:msg]))

		// Broadcast the message to all clients
		s.broadcast(client.name + ": " + string(buffer[:msg]))
	}

}

func main() {

	server := "localhost"
	port := "8080"
	server_instance := Server{make(map[string]Client)}

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
		go server_instance.handleConnection(conn)
	}

}
