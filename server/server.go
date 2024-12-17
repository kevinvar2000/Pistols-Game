package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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

	// Read the first message for player registration
	message, err := read_message(conn)
	if err != nil {
		fmt.Println("Error reading from client:", err.Error())
		return
	}

	// Parse the player name from the message
	if !strings.HasPrefix(message, "name:") {
		conn.Write([]byte("response_type=error&status_code=400&message=Player name required\n"))
		return
	}
	player_name := strings.TrimPrefix(message, "name:")
	player_name = strings.TrimSpace(player_name)

	// Create player
	player := create_player(conn, player_name)

	// Wait for other players to join
	wait_for_players(player)

	for {

		// Read message from client
		message, err := read_message(conn)
		if err != nil {
			fmt.Println("Error reading from client:", err.Error())
			break
		}

		message = strings.TrimSpace(message)

		// Handle different types of requests
		switch {
		case message == "exit":
			exit_game(player)
		case strings.HasPrefix(message, "action:"):
			set_player_action(player, strings.ToUpper(strings.TrimPrefix(message, "action:")))
		case message == "is_alive":
			is_alive(conn)
		default:
			invalid_message(player)
		}
	}
}

func invalid_message(player *Player) {
	player.conn.Write([]byte("response_type=error&status_code=400&message=Invalid request\n"))
}

func read_message(conn net.Conn) (string, error) {

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", err
	}
	return string(buffer[:n]), nil
}

func is_alive(conn net.Conn) {
	conn.Write([]byte("response_type=alive&status_code=200&message=I am alive\n"))
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
