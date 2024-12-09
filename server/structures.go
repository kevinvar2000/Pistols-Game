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
	clients  []*Client
	mu       sync.Mutex
	states   map[*Client]*PlayerState
	gameOver bool
}

// GameServer structure
type GameServer struct {
	rooms []*Room
	mu    sync.Mutex
}

// PlayerState structure
type PlayerState struct {
	health int
	ammo   int
	action string
}
