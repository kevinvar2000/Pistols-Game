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
		// fmt.Println("Received message:", message)

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
		request := strings.TrimPrefix(message, "request_type=")

		if strings.HasPrefix(request, "ping") {
			ping(conn)
			continue
		}

		if is_registered {

			fmt.Printf("Player %s sent request: %s\n", player.name, request)

			switch {
			case strings.HasPrefix(request, "game_ready"):
				game_ready(player)
			case strings.HasPrefix(request, "player_state"):
				get_player_state(player)
			case strings.HasPrefix(request, "exit_game"):
				exit_game(player)
			case strings.HasPrefix(request, "action"):
				set_player_action(player, request)
				conn.Write([]byte("response_type=action&status_code=200&message=Action set\n"))
			case strings.HasPrefix(request, "game_result"):
				get_game_result(player)
			case strings.HasPrefix(request, "round_state"):
				get_round_state(player)
			default:
				invalid_message(player)
			}
		} else {
			// If the player is not registered, allow only "name:" message for registration
			if strings.HasPrefix(request, "name") {
				player = register_player(conn, request)
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

func register_player(conn net.Conn, request string) *Player {

	player_name := strings.TrimPrefix(request, "name&name=")
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
