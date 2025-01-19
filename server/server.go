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
			case request == "start_game":

				if player == nil {
					fmt.Println("Player is nil.")
					conn.Write([]byte("response_type=start_game&status_code=403&message=Player not registered\n"))
					continue
				}

				// Check if the player is already in a game
				if player.game != nil {

					if player.game.game_state == "running" {
						conn.Write([]byte("response_type=start_game&status_code=403&message=Game already running\n"))
						continue
					}

				}

				if player.player_state.is_dead {
					conn.Write([]byte("response_type=start_game&status_code=403&message=Player is dead, send reset_game\n"))
					continue
				}

				go wait_for_players(player)

			case request == "game_state":
				get_game_state_req(player)
			case request == "close_game":
				close_game(player)
				return
			case request == "reset_game":
				reset_game(player)
			case strings.HasPrefix(request, "action"):
				set_player_action(player, request)
			case strings.HasPrefix(request, "name"):
				conn.Write([]byte("response_type=error&status_code=403&message=Player already registered\n"))
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

			} else {
				// invalid_message(conn)
				conn.Write([]byte("response_type=error&status_code=403&message=Player not registered\n"))
			}
		}
	}
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
