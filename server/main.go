package main

import (
	"fmt"
	"net"
	"os"
)

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
		go server.Handle_client(client)
	}
}
