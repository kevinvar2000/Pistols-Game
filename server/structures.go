package main

import (
	"net"
	"sync"
)

// Client structure
type Client struct {
	conn net.Conn
	name string
	room *Room
}

// Room structure
type Room struct {
	clients []*Client
	mu      *sync.Mutex
}

// GameServer manages the connected clients and rooms
type GameServer struct {
	rooms []*Room
	mu    sync.Mutex
}
