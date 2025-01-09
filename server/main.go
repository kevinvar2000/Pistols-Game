package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	// Get the server address from command line arguments or use default
	address := get_address()
	// Start the server and get the listener
	listener := start_server(address)
	// Ensure the listener is closed when the main function exits
	defer listener.Close()

	// Accept incoming connections
	accept_connections(listener)
}

// get_address retrieves the server address from command line arguments
// or falls back to a default address if none is provided or if the provided address is invalid
func get_address() string {
	args := os.Args
	if len(args) >= 2 {
		address := args[1]
		if is_valid_address(address) {
			return address
		}
		fmt.Println("Invalid address provided, falling back to default.")
	}
	return CONN_ADDRESS + ":" + CONN_PORT
}

// is_valid_address checks if the provided address is a valid IP address and port
func is_valid_address(address string) bool {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	if net.ParseIP(host) == nil {
		fmt.Println("Invalid IP address:", host)
		return false
	}

	portInt, err := strconv.Atoi(port)
	if err != nil || portInt < 1 || portInt > 65535 {
		fmt.Println("Invalid port:", port)
		return false
	}

	return true
}
