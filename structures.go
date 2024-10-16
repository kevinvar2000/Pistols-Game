package main

import "net"

// Client structure
type Client struct {
	conn net.Conn
	name string
}

// Server structure
type Server struct {
	clients map[string]Client
}
