package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// Remove a client from the room
func (r *Room) removeClient(client *Client, server *GameServer) {

	fmt.Println("***Removing client from room***")

	if r == nil {
		fmt.Println("Warning: Attempted to remove client from a nil room")
		return
	}

	r.mu.Lock()
	fmt.Println("Locking room in removeClient")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in removeClient")

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
		r.broadcast(fmt.Sprintf("Player %s has left the room.", client.name), client)
		r.mu.Lock()
	} else {
		// If no clients remain, clean up the room
		fmt.Println("Room is now empty; cleaning up.")
		client.room = nil

		// Perform cleanup after unlocking to avoid deadlocks
		go func() {
			server.mu.Lock()
			defer server.mu.Unlock()
			server.removeRoom(r)
		}()
	}
	fmt.Println()
}

func (s *GameServer) removeRoom(room *Room) {

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
func (r *Room) broadcast(message string, sender *Client) {

	fmt.Println("***Broadcasting message to all clients in the room***")

	r.mu.Lock()
	fmt.Println("Locking room in broadcast")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in broadcast")

	for _, client := range r.clients {
		if client != sender {
			fmt.Fprintf(client.conn, message+"\n")
		}
	}
	fmt.Println("Broadcast message:", message)
	fmt.Println()
}

// Handle client communication and room assignment
func (s *GameServer) handleClient(client *Client) {

	fmt.Println("***Handling client***")

	defer func() {
		// Handle client disconnection or early exit
		if client.room != nil && len(client.room.clients) == 2 {
			fmt.Println("Client disconnected:", client.name)
			client.room.removeClient(client, s)
			client.room.broadcast(fmt.Sprintf("%s has left the game.", client.name), client) // Inform the other player
		} else {
			fmt.Println("Client left before game started:", client.name)
		}
		client.conn.Close()
	}()

	// **Register client name**
	if err := s.registerClientName(client); err != nil {
		fmt.Println("Error registering client:", err)
		return
	}

	// **Wait for the second player to join**
	if err := s.waitForSecondPlayer(client); err != nil {
		fmt.Println("Error waiting for second player:", err)
		return
	}

	// **Start the game**
	if err := s.startGame(client); err != nil {
		fmt.Println("Error starting the game:", err)
		return
	}

	// **Handle player actions during the game**
	s.handleGameActions(client)

	// Start the ping goroutine to monitor the client's activity
	go s.pingClient(client)
	fmt.Println()
}

// Ping the client to check if they are still alive
func (s *GameServer) pingClient(client *Client) {

	fmt.Println("***Pinging client***")

	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)) // Set timeout for ping response
		_, err := fmt.Fprintf(client.conn, "PING\n")
		if err != nil {
			fmt.Println("Ping failed for client", client.name, "Error:", err)
			client.room.removeClient(client, s) // Remove client from room if ping fails
			return
		}
	}
	fmt.Println()
}

func (s *GameServer) registerClientName(client *Client) error {

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

	client.room = s.assignRoom(client)
	fmt.Println("Assigned room to client:", client.name, "Room:", client.room)

	fmt.Println()
	return nil
}

// Step 2: Wait for the second player to join
func (s *GameServer) waitForSecondPlayer(client *Client) error {

	fmt.Println("***Waiting for second player to join***")

	client.conn.Write([]byte("WAITING_FOR_PLAYER\n"))
	fmt.Println("Waiting for another player to join the room:", client.room.clients[0].name)

	timeout := time.After(30 * time.Second) // Set a 30-second timeout
	fmt.Println("Timeout set to 30 seconds")

	for {
		if client.room == nil {
			return fmt.Errorf("client.room is nil")
		}

		client.room.mu.Lock()
		currentCount := len(client.room.clients)
		// fmt.Printf("Current room count: %d\n", currentCount)

		if currentCount == 2 {
			client.room.mu.Unlock()
			fmt.Println("Both players have joined. Proceeding with the game.")
			break
		}
		client.room.mu.Unlock()

		select {
		case <-timeout:
			fmt.Println("Timeout reached: No second player joined.")
			client.conn.Write([]byte("TIMEOUT: No second player joined.\n"))
			client.room.removeClient(client, s)
			return fmt.Errorf("timeout reached: no second player joined")
		default:
			time.Sleep(100 * time.Millisecond) // Prevent busy looping
		}
	}
	fmt.Println()
	return nil
}

// Step 3: Start the game
func (s *GameServer) startGame(client *Client) error {

	fmt.Println("***Starting the game***")

	client.conn.Write([]byte("STARTING_GAME\n"))
	client.room.broadcast(fmt.Sprintf("%s has joined. Game starts now!", client.name), client)

	fmt.Println()
	return nil
}

// Step 4: Handle player actions during the game
func (s *GameServer) handleGameActions(client *Client) {

	fmt.Println("***Handling player actions during the game***")

	reader := bufio.NewReader(client.conn)
	client.conn.Write([]byte("GAME: Enter your action (SHOOT, COVER, RELOAD) or type DISCONNECT to leave: \n"))

	for {
		// Read player action
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection:", client.name)
			}
			fmt.Printf("Error reading message from client %s: %v\n", client.name, err)
			client.room.removeClient(client, s)
			break
		}

		message = strings.TrimSpace(message)
		fmt.Println("Received message from client:", message)

		message = strings.ToUpper(message)

		// Check for correct message format
		if message != "SHOOT" && message != "COVER" && message != "RELOAD" && message != "DISCONNECT" {
			client.conn.Write([]byte("GAME: Invalid action. Enter SHOOT, COVER, RELOAD, or DISCONNECT: \n"))
			continue
		}

		// Handle client disconnection
		if message == "DISCONNECT" {
			client.room.removeClient(client, s)
			break
		}

		// **Handle player actions**
		// if strings.HasPrefix(message, "PLAYER_ACTION:") {
		// 	client.room.handleAction(client, strings.TrimPrefix(message, "PLAYER_ACTION:"))
		// }
		client.room.handleAction(client, message)

		client.conn.Write([]byte("GAME: Enter your action (SHOOT, COVER, RELOAD) or type DISCONNECT to leave: \n"))

	}
	fmt.Println()
}

func (r *Room) handleAction(client *Client, action string) {

	fmt.Println("***Handling player action in the room***")

	r.mu.Lock()
	fmt.Println("Locking room in handleAction")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in handleAction")

	state := r.states[client]
	state.action = strings.ToUpper(action)
	fmt.Println("Client:", client.name, "Action:", state.action)

	if len(r.clients) == 2 && r.states[r.clients[0]].action != "" && r.states[r.clients[1]].action != "" {
		// Resolve actions
		player1 := r.states[r.clients[0]]
		player2 := r.states[r.clients[1]]

		r.resolveActions(player1, player2)
		r.resetActions()
	}
	fmt.Println()
}

func (r *Room) resolveActions(p1, p2 *PlayerState) {

	fmt.Println("***Resolving player actions in the room***")

	// Lock the room before making changes
	r.mu.Lock()
	fmt.Println("Locking room in resolveActions")
	defer r.mu.Unlock()
	fmt.Println("Unlocking room in resolveActions")

	fmt.Println("Resolving actions:", p1.action, p2.action)
	if p1.action == "SHOOT" && p2.action != "COVER" {
		fmt.Println("Player 1 shoots Player 2.")
		p2.health -= 1
	}
	if p2.action == "SHOOT" && p1.action != "COVER" {
		fmt.Println("Player 2 shoots Player 1.")
		p1.health -= 1
	}

	if p1.action == "RELOAD" {
		fmt.Println("Player 1 reloads.")
		p1.ammo += 1
	}

	if p2.action == "RELOAD" {
		fmt.Println("Player 2 reloads.")
		p2.ammo += 1
	}

	var result_msg string
	if p1.health <= 0 && p2.health <= 0 {
		fmt.Println("Both players are dead.")
		result_msg = "RESULT: Both Players are Dead!"
	} else if p1.health <= 0 {
		fmt.Println("Player 2 wins.")
		result_msg = "RESULT: Player 2 Wins!"
	} else if p2.health <= 0 {
		fmt.Println("Player 1 wins.")
		result_msg = "RESULT: Player 1 Wins!"
	} else {
		fmt.Println("Players are still alive.")
		result_msg = fmt.Sprintf("RESULT: Player1 - %d HP, Player2 - %d HP", p1.health, p2.health)
	}

	r.broadcast(result_msg, nil)
	fmt.Println()
}

func (r *Room) resetActions() {

	fmt.Println("***Resetting actions for all clients in the room***")

	for _, client := range r.clients {
		r.states[client].action = ""
	}
	fmt.Println("Actions reset for all clients in the room")
	fmt.Println()
}

// Create or find a room for a new client
func (s *GameServer) assignRoom(client *Client) *Room {

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

func main() {
	server := &GameServer{
		rooms: make([]*Room, 0),
	}

	var address string

	// Read arguments from the command line
	args := os.Args
	if len(args) >= 2 {
		address = args[1]
	} else {
		address = CONN_ADDRESS + ":" + CONN_PORT
	}

	// Listen for incoming connections
	listener, err := net.Listen(CONN_NETWORK, address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Starting server on", address)

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
