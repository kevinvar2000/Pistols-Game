package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// start_server initializes a server listener on the given address
func start_server(address string) net.Listener {
	listener, err := net.Listen(CONN_NETWORK, address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Starting server on", address)
	return listener
}

// accept_connections continuously accepts incoming connections
func accept_connections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr().String())

		// Handle each connection in a separate goroutine
		go handle_connection(conn)
	}
}

// handle_connection processes messages from a single connection
func handle_connection(conn net.Conn) {
	defer conn.Close()

	hello_message := false
	is_registered := false
	var player *Player

	for {
		// Set a read deadline for the connection
		conn.SetReadDeadline(time.Now().Add(READ_DEADLINE * time.Second))

		// Read the first message for player registration
		message, err := read_message(conn)
		if err != nil {
			// Handle error reading from client
			if err == io.EOF {
				fmt.Println("Client disconnected.")
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Connection timed out.")
				conn.Write([]byte("response_type=error&status_code=408&message=Connection timed out\n"))
			} else {
				fmt.Println("Error reading from client:", err.Error())
			}

			// If the player was registered, handle disconnection
			if is_registered && player != nil {
				fmt.Println("Calling close_game for player:", player.name)
				close_game(player)
			}

			break
		}

		// Trim the message of any leading or trailing spaces
		message = strings.TrimSpace(message)

		// Check if the message is empty
		if len(message) == 0 {
			conn.Write([]byte("response_type=error&status_code=400&message=Empty message received\n"))
			continue
		}

		// Validate the message format
		if !strings.HasPrefix(message, "request_type") {
			invalid_message(conn)
			handle_disconnection(player)
			return
		}

		// Trim the message of the request_type prefix
		request := strings.TrimPrefix(message, "request_type=")

		// Handle ping requests
		if request == "ping" {
			ping(conn)
			continue
		}

		// Check if the first message is a hello message
		if !hello_message {

			fmt.Println("First message:", request)

			// Check if the first message is a hello message
			if request != "hello" {
				fmt.Println("Invalid first message.")
				conn.Write([]byte("response_type=hello&status_code=400&message=Invalid first message\n"))
				handle_disconnection(player)
				return
			} else {
				fmt.Println("Hello message received.")
				conn.Write([]byte("response_type=hello&status_code=200&message=Hello\n"))
				hello_message = true
				continue
			}
		}

		// Handle hello messages if sent after the first message
		if request == "hello" {
			fmt.Println("Hello message received again.")
			conn.Write([]byte("response_type=hello&status_code=200&message=Hello again\n"))
			continue
		}

		// Handle registered player requests
		if is_registered {
			fmt.Printf("Player %s sent request: %s\n", player.name, request)

			switch {
			case request == "game_ready":
				game_ready(player)
			case request == "player_state":
				get_player_state(player)
			case request == "opponent_state":
				get_opponent_state(player)
			case request == "close_game":
				close_game(player)
				return
			case request == "reset_game":
				reset_game(player)
			case strings.HasPrefix(request, "action"):
				set_player_action(player, request)
			case request == "game_result":
				get_game_result(player)
			case request == "round_state":
				get_round_state(player)
			case request == "game_state":
				get_game_state(player)
			default:
				invalid_message(conn)
				handle_disconnection(player)
				return
			}
		} else {
			// If the player is not registered, allow only "name:" message for registration
			if strings.HasPrefix(request, "name") {
				player = register_player(conn, request)

				// If the player is nil, return
				if player == nil {
					continue
				}

				is_registered = true

				go wait_for_players(player)
			} else {
				// invalid_message(conn)
				conn.Write([]byte("response_type=error&status_code=403&message=Player not registered\n"))
			}
		}
	}
}

// register_player registers a new player with the given request
func register_player(conn net.Conn, request string) *Player {

	// Check if the request is valid
	if !strings.HasPrefix(request, "name&name=") {
		fmt.Println("Invalid name request.")
		conn.Write([]byte("response_type=error&status_code=400&message=Invalid name request\n"))
		return nil
	}

	// Extract the player name from the request
	player_name := strings.TrimPrefix(request, "name&name=")

	// Check if the player name is empty
	if len(player_name) == 0 || player_name == "" {
		fmt.Println("Empty player name received.")
		conn.Write([]byte("response_type=error&status_code=400&message=Empty player name\n"))
		return nil
	}

	// Check if the player name is too long
	if len(player_name) > MAX_PLAYER_NAME_LENGTH {
		fmt.Println("Player name too long.")
		conn.Write([]byte("response_type=error&status_code=400&message=Player name too long\n"))
		return nil
	}

	fmt.Println("Registering player:", player_name)

	var new_player *Player

	if player, exists := disconnected_players[player_name]; exists {
		fmt.Println("Player reconnected:", player_name)

		// Reconnect the player
		new_player = player
		new_player.conn = conn

		// Update the game's state if necessary
		player.game.mutex.Lock()
		player.game.players[player] = true // Ensure the player is re-added to the game's player list
		player.game.mutex.Unlock()

		// Remove the player from the disconnected players list
		delete(disconnected_players, player_name)

		// Send a response to the player
		conn.Write([]byte("response_type=name&status_code=200&message=Player reconnected\n"))

	} else {
		fmt.Println("New player registered:", player_name)

		new_player = create_player(conn, player_name)

		// Send a response to the player
		conn.Write([]byte("response_type=name&status_code=200&message=Player registered\n"))
	}

	return new_player
}

// handle_disconnection of a player
func handle_disconnection(player *Player) {

	// Remove the player from the game
	if player != nil {
		fmt.Println("Calling close_game for player:", player.name)
		close_game(player)
	}
}

// invalid_message sends an error response for invalid messages
func invalid_message(conn net.Conn) {
	fmt.Println("Invalid message received.")
	conn.Write([]byte("response_type=error&status_code=400&message=Invalid request\n"))
}

// read_message reads a message from the connection
func read_message(conn net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return "", err
	}
	return string(buffer[:n]), nil
}

// ping sends a pong response to the client
func ping(conn net.Conn) {
	fmt.Println("Pinging player...")
	conn.Write([]byte("response_type=ping&status_code=200&message=Pong\n"))
}
