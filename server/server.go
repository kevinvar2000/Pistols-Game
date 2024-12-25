package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

func start_server(address string) net.Listener {
	listener, err := net.Listen(CONN_NETWORK, address)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Starting server on", address)
	return listener
}

func accept_connections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr().String())

		go handle_connection(conn)
	}
}

func handle_connection(conn net.Conn) {
	defer conn.Close()

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
			} else {
				fmt.Println("Error reading from client:", err.Error())
			}
			break
		}

		// Trim the message of any leading or trailing spaces
		message = strings.TrimSpace(message)
		fmt.Println("Received message:", message)

		// Check if the message is empty
		if len(message) == 0 {
			conn.Write([]byte("response_type=error&status_code=400&message=Empty message received\n"))
			continue
		}

		if !strings.HasPrefix(message, "request_type") {
			invalid_message(player)
			continue
		}

		// Trim the message of the request_type prefix
		message = strings.TrimPrefix(message, "request_type=")

		if strings.HasPrefix(message, "ping") {
			ping(conn)
			continue
		}

		if is_registered {

			fmt.Println("Message received from registered player:", player.name)
			fmt.Println("Message:", message)

			switch {
			case strings.HasPrefix(message, "game_ready"):
				game_ready(player)
			case strings.HasPrefix(message, "player_state"):
				get_player_state(player)
			case strings.HasPrefix(message, "exit_game"):
				exit_game(player)
			case strings.HasPrefix(message, "action"):
				set_player_action(player, message)
				conn.Write([]byte("response_type=action&status_code=200&message=Action set\n"))
			default:
				invalid_message(player)
			}
		} else {
			// If the player is not registered, allow only "name:" message for registration
			if strings.HasPrefix(message, "name") {
				player = register_player(conn, message)
				is_registered = true

				go wait_for_players(player)
			} else {
				if player != nil {
					invalid_message(player)
				}
			}
		}
	}

}

func register_player(conn net.Conn, message string) *Player {

	player_name := strings.TrimPrefix(message, "name&name=")
	fmt.Println("Registering player:", player_name)
	conn.Write([]byte("response_type=name&status_code=200&message=Player registered\n"))

	return create_player(conn, player_name)
}

func invalid_message(player *Player) {
	fmt.Println("Invalid message received.")
	player.conn.Write([]byte("response_type=error&status_code=400&message=Invalid request\n"))
}

func read_message(conn net.Conn) (string, error) {

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return "", err
	}
	return string(buffer[:n]), nil
}

func ping(conn net.Conn) {

	fmt.Println("Pinging player...")
	conn.Write([]byte("response_type=ping&status_code=200&message=Pong\n"))
}

func broadcast_message(game *Game, message string) {

	game.mutex.Lock()
	for player := range game.players {

		if player.conn == nil {
			continue
		}

		player.conn.Write([]byte(message))
	}
	game.mutex.Unlock()
}
