package main

import (
	"net"
	"sync"
)

// Game represents the state of a game session.
type Game struct {
	players     map[*Player]bool // A map of players in the game.
	game_state  string           // The current state of the game.
	round_state string           // The current state of the round.
	mutex       *sync.Mutex      // Mutex to protect game state.
}

// Player represents a player in the game.
type Player struct {
	conn         net.Conn    // Network connection of the player.
	name         string      // Name of the player.
	player_state PlayerState // Current state of the player.
	game         *Game       // Reference to the game the player is in.
	mutex        sync.Mutex  // Mutex to protect player state.
}

// PlayerState represents the state of a player.
type PlayerState struct {
	health  int    // Health of the player.
	ammo    int    // Ammo count of the player.
	action  string // Current action of the player.
	is_dead bool   // Whether the player is dead.
}
