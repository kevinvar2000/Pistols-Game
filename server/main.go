package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	address := get_address()
	listener := start_server(address)
	defer listener.Close()

	accept_connections(listener)
}

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
