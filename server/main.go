package main

import (
	"os"
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
		return args[1]
	}
	return CONN_ADDRESS + ":" + CONN_PORT
}
