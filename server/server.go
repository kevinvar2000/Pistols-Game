package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Add a client to the room
func (r *Room) addClient(client *Client) bool {

	if r == nil {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.clients) < 2 {
		r.clients = append(r.clients, client)
		return true
	}
	return false
}

// Remove a client from the room
func (r *Room) removeClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, c := range r.clients {
		if c == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			break
		}
	}
}

// Broadcast sends a message to all clients in the room
func (r *Room) broadcast(message string, sender *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clients {
		if client != sender {
			fmt.Fprintf(client.conn, message+"\n")
		}
	}
}

// Handle client communication and room assignment
func (s *GameServer) handleClient(client *Client) {
	defer client.conn.Close()

	reader := bufio.NewReader(client.conn)
	client.conn.Write([]byte("Welcome to the game! Enter your name: \n"))

	// First message from client is their name
	name, _ := reader.ReadString('\n')
	client.name = strings.TrimSpace(name)

	client.room = s.assignRoom(client)
	client.conn.Write([]byte("You have been placed in a room. Waiting for an opponent...\n"))

	// Wait until the room is full
	for {
		client.room.mu.Lock()
		if len(client.room.clients) == 2 {
			client.room.mu.Unlock()
			break
		}
		client.room.mu.Unlock()
	}

	client.conn.Write([]byte("Game starts now!\n"))
	client.room.broadcast(fmt.Sprintf("%s has joined the game!", client.name), client)

	for {
		// Read messages from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%s has disconnected.\n", client.name)
			client.room.removeClient(client)
			client.room.broadcast(fmt.Sprintf("%s has left the game.", client.name), client)
			break
		}

		message = strings.TrimSpace(message)
		if message == "quit" {
			client.room.broadcast(fmt.Sprintf("%s has left the game.", client.name), client)
			client.room.removeClient(client)
			break
		}

		// Broadcast the client's message to their room
		client.room.broadcast(fmt.Sprintf("%s: %s", client.name, message), client)
	}
}

// Create or find a room for a new client
func (s *GameServer) assignRoom(client *Client) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range s.rooms {
		if room.addClient(client) {
			return room
		}
	}

	// If no available room, create a new one
	newRoom := &Room{
		clients: []*Client{},
		mu:      &sync.Mutex{},
	}
	newRoom.addClient(client)
	s.rooms = append(s.rooms, newRoom)
	return newRoom
}

func main() {

	server := &GameServer{}

	address := "0.0.0.0"
	port := "8080"
	// server_instance := Server{make(map[string]Client)}

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
	listener, err := net.Listen("tcp", address+":"+port)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Starting server on", address+":"+port)

	// Accept connections in a loop
	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}

		fmt.Println("Accepted connection from", conn.RemoteAddr().String())

		client := &Client{conn: conn}
		go server.handleClient(client)
	}

}
