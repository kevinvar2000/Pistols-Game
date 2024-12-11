package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// Remove a client from the room
func (r *Room) Remove_client(client *Client, server *GameServer) {

	fmt.Println("***Removing client from room***")

	if r == nil {
		fmt.Println("Warning: Attempted to remove client from a nil room")
		return
	}

	r.mu.Lock()
	fmt.Println("Locking room in Remove_client")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in Remove_client")

	for i, c := range r.clients {
		if c == client {
			r.clients = append(r.clients[:i], r.clients[i+1:]...)
			fmt.Println("Client removed from room:", client.name)
			break
		}
	}

	println("Remaining clients in the room:", len(r.clients))

	// Notify remaining clients
	if len(r.clients) > 0 {
		fmt.Println("Notifying remaining clients in the room.")
		r.mu.Unlock()
		r.Broadcast(fmt.Sprintf("Player %s has left the room.", client.name), client)
		r.mu.Lock()
	} else {
		// If no clients remain, clean up the room
		fmt.Println("Room is now empty; cleaning up.")
		client.room = nil

		// Perform cleanup after unlocking to avoid deadlocks
		go func() {
			server.mu.Lock()
			defer server.mu.Unlock()
			server.Remove_room(r)
		}()
	}
	fmt.Println()
}

// Create or find a room for a new client
func (s *GameServer) Assign_room(client *Client) *Room {

	fmt.Println("***Assigning room to client***")

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range s.rooms {

		if len(room.clients) < 2 {
			room.clients = append(room.clients, client)

			if room.states == nil {
				room.states = make(map[*Client]*PlayerState)
			}

			room.states[client] = &PlayerState{health: 3, ammo: 1, action: "COVER"}
			fmt.Println("Client assigned to existing room:", client.name)
			fmt.Println()
			return room
		}
	}

	newRoom := &Room{
		clients: []*Client{client},
		states:  make(map[*Client]*PlayerState),
	}
	fmt.Println("New room created:", newRoom)

	newRoom.states[client] = &PlayerState{health: 3, ammo: 1, action: "COVER"}
	fmt.Println("Client state set in new room:", client.name, newRoom.states[client])

	fmt.Println("Appending new room to server:", newRoom)
	s.rooms = append(s.rooms, newRoom)
	fmt.Println("Client assigned to new room:", client.name)

	fmt.Println()
	return newRoom
}

func (s *GameServer) Remove_room(room *Room) {

	fmt.Println("***Removing empty room from server***")

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.rooms {
		if r == room {
			s.rooms = append(s.rooms[:i], s.rooms[i+1:]...)
			fmt.Println("Removed empty room from server.")
			break
		}
	}
	fmt.Println()
}

// Broadcast sends a message to all clients in the room
func (r *Room) Broadcast(message string, sender *Client) {

	fmt.Println("***Broadcasting message to all clients in the room***")

	r.mu.Lock()
	fmt.Println("Locking room in Broadcast")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in Broadcast")

	for _, client := range r.clients {
		if client != sender {
			fmt.Fprintf(client.conn, message+"\n")
		}
	}
	fmt.Println("Broadcast message:", message)
	fmt.Println()
}

// Handle client communication and room assignment
func (s *GameServer) Handle_client(client *Client) {

	fmt.Println("***Handling client***")

	defer func() {
		// Handle client disconnection or early exit
		if client.room != nil && len(client.room.clients) == 2 {
			fmt.Println("Client disconnected:", client.name)
			client.room.Remove_client(client, s)
			client.room.Broadcast(fmt.Sprintf("%s has left the game.", client.name), client) // Inform the other player
		} else {
			fmt.Println("Client left before game started:", client.name)
		}
		client.conn.Close()
	}()

	// **Register client name**
	if err := s.Register_client_name(client); err != nil {
		fmt.Println("Error registering client:", err)
		return
	}

	// **Wait for the second player to join**
	if err := s.Wait_for_player(client); err != nil {
		fmt.Println("Error waiting for second player:", err)
		return
	}

	// **Start the game**
	if err := s.Start_game(client); err != nil {
		fmt.Println("Error starting the game:", err)
		return
	}

	// **Handle player actions during the game**
	s.Handle_game_actions(client)

	// Start the ping goroutine to monitor the client's activity
	go s.Ping_client(client)
	fmt.Println()
}

// Ping the client to check if they are still alive
func (s *GameServer) Ping_client(client *Client) {

	fmt.Println("***Pinging client***")

	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)) // Set timeout for ping response
		_, err := fmt.Fprintf(client.conn, "PING\n")
		if err != nil {
			fmt.Println("Ping failed for client", client.name, "Error:", err)
			client.room.Remove_client(client, s) // Remove client from room if ping fails
			return
		}
	}
	fmt.Println()
}

func (s *GameServer) Register_client_name(client *Client) error {

	fmt.Println("***Registering client name***")

	reader := bufio.NewReader(client.conn)
	client.conn.Write([]byte("CONNECT: Enter your name: \n"))

	// Read name from client
	name, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read name: %v", err)
	}

	client.name = strings.TrimSpace(name)
	client.conn.Write([]byte(fmt.Sprintf("CONNECTING: %s\n", client.name)))
	fmt.Println("Client connected with name:", client.name)

	client.room = s.Assign_room(client)
	fmt.Println("Assigned room to client:", client.name, "Room:", client.room)

	fmt.Println()
	return nil
}
