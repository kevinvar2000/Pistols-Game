package main

import (
	"net"
	"sync"
)

type Game struct {
	players    map[*Player]bool
	game_state string
	mutex      *sync.Mutex
}

type Player struct {
	conn         net.Conn
	name         string
	player_state PlayerState
	game         *Game
	mutex        sync.Mutex
}

type PlayerState struct {
	health int
	ammo   int
	action string
}
