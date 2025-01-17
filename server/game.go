package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var waiting_players = make(chan *Player, MAX_PLAYERS)
var disconnected_players = make(map[string]*Player)

// Function to wait for players to join the game
func wait_for_players(player *Player) {
	fmt.Println("Waiting for players to join...")

	var game *Game

	// Check if the player is reconnecting
	if player.game != nil {

		fmt.Println("Players game is not nil...")

		player.game.mutex.Lock()
		if player.game.game_state == "reconnect" {

			fmt.Println("Player", player.name, "is reconnecting...")
			game = player.game
			player.game.game_state = "running"
			player.game.mutex.Unlock()

			fmt.Println("Reconnecting player to the game...")
			game.start_game()
			return
		} else {
			fmt.Println("Player's game state is not reconnect.")
		}

		player.game.mutex.Unlock()
	} else {
		fmt.Println("Player's game is nil.")
	}

	fmt.Println("Adding player to the waiting list...")

	// Add player to the waiting list
	waiting_players <- player

	if len(waiting_players) == MAX_PLAYERS {

		// Check if the player's existing game is reusable
		if player.game != nil {
			player.game.mutex.Lock()
			if player.game.game_state == "waiting" {
				fmt.Println("Reusing the existing game...")
				game = player.game
			}
			player.game.mutex.Unlock()
		}

		// If no reusable game, create a new one
		if game == nil {
			game = create_game()
		}

		// Add players to the game
		for i := 0; i < MAX_PLAYERS; i++ {
			player := <-waiting_players
			game.players[player] = true
			player.game = game
			fmt.Println("Player joined the game:", player.name)
		}

		// drain waiting players
		for len(waiting_players) > 0 {
			<-waiting_players
		}

		// Start the game
		game.start_game()
	}
}

// Function to check if the game is ready
func game_ready(player *Player) {
	if player.game != nil {
		if player.game.game_state == "running" {
			fmt.Println("Game is ready.")
			player.conn.Write([]byte("response_type=game_ready&status_code=200&message=Game ready\n"))
		} else if player.game.game_state == "waiting" {
			fmt.Println("Game is in waiting state.")
			player.conn.Write([]byte("response_type=game_ready&status_code=200&message=Game waiting\n"))
		} else if player.game.game_state == "reconnect" {
			fmt.Println("Game is waiting for player to reconnect.")
			player.conn.Write([]byte("response_type=game_ready&status_code=200&message=Game reconnect\n"))
		} else {
			fmt.Println("Game is over.")
			player.conn.Write([]byte("response_type=game_ready&status_code=409&message=Game over\n"))
		}
	} else {
		fmt.Println("Game is not ready.")
		player.conn.Write([]byte("response_type=game_ready&status_code=400&message=Game not ready\n"))
	}
}

// Function to create a new player
func create_player(conn net.Conn, player_name string) *Player {
	fmt.Println("Creating player:", player_name)

	player := &Player{
		conn:         conn,
		player_state: PlayerState{health: DEFAULT_HEALTH, ammo: DEFAULT_AMMO, action: ""},
	}
	player.name = player_name
	return player
}

// Function to create a new game
func create_game() *Game {
	fmt.Println("Creating a new game...")

	game := &Game{
		players:    make(map[*Player]bool),
		game_state: "waiting",
		mutex:      &sync.Mutex{},
	}
	return game
}

// Method to start the game
func (game *Game) start_game() {
	fmt.Println("Starting the game...")

	game.game_state = "running"

	// Start the game loop
	go game.game_loop()
}

// Method to run the game loop
func (game *Game) game_loop() {
	fmt.Println("Game loop started...")

	for {
		// Check if the game is running
		if game.game_state != "running" {
			fmt.Println("Game loop is not running...")
			break
		}

		if game.wait_for_all_actions() {
			// Check player actions
			game.check_player_actions()

			// Reset player actions
			game.reset_player_actions()

			// Check player health
			game.check_player_health()
		}

		time.Sleep(time.Duration(ROUND_WAIT_TIME) * time.Second)
	}

	fmt.Println("Game loop ended...")
}

// Method to wait for all players to perform actions
func (game *Game) wait_for_all_actions() bool {
	fmt.Println("Waiting for player actions...")

	game.mutex.Lock()
	game.round_state = "running"
	game.mutex.Unlock()

	for {
		all_ready := true

		game.mutex.Lock()

		// Handle "reconnect" state
		if game.game_state == "reconnect" {
			fmt.Println("Game state is reconnect in wait for all actions.")
			game.mutex.Unlock()

			// Wait for reconnect resolution before proceeding
			time.Sleep(1 * time.Second)
			continue
		}

		// Check if the game is running
		if game.game_state != "running" {
			fmt.Println("Game state is not running in wait for all actions.")
			game.mutex.Unlock()
			return false
		}

		// Check if all players have performed actions
		for player := range game.players {
			player.mutex.Lock()
			if player.player_state.action == "" {
				all_ready = false
			}
			player.mutex.Unlock()
		}
		game.mutex.Unlock()

		if all_ready {
			fmt.Println("All players are ready!")
			game.mutex.Lock()
			game.round_state = "end"
			game.mutex.Unlock()
			return true
		}

		time.Sleep(time.Duration(ROUND_WAIT_TIME) * time.Second)

	}
}

// Function to set the player's action
func set_player_action(player *Player, message string) {

	// Check if the message is valid
	if !strings.HasPrefix(message, "action&action_type=") {
		fmt.Println("Invalid action request.")
		player.conn.Write([]byte("response_type=action&status_code=400&message=Invalid action request\n"))
		return
	}

	// action&action_type=shoot
	action := strings.TrimPrefix(message, "action&action_type=")

	fmt.Println("Setting player action:", action)

	// Set the player action
	player.mutex.Lock()
	defer player.mutex.Unlock()

	// Check if player is registered in a game
	if player.game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=action&status_code=403&message=Player not in game\n"))
		return
	}

	// Check if game is running
	if player.game.game_state != "running" {
		fmt.Println("Game is not running.")
		player.conn.Write([]byte("response_type=action&status_code=409&message=Game not running\n"))
		return
	}

	// Check if player is dead
	if player.player_state.action != "" {
		fmt.Println("Player action already set.")
		player.conn.Write([]byte("response_type=action&status_code=409&message=Action already set\n"))
		return
	}

	action = strings.ToUpper(action)

	// Check if the action is valid
	if action != ACTION_SHOOT && action != ACTION_COVER && action != ACTION_RELOAD {
		fmt.Println("Invalid action:", action)
		player.conn.Write([]byte("response_type=action&status_code=400&message=Invalid action\n"))
		return
	}

	// Set the player action
	player.player_state.action = action
	player.conn.Write([]byte("response_type=action&status_code=200&message=Action set\n"))
}

// Method to check player actions
func (game *Game) check_player_actions() {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	for player := range game.players {
		switch player.player_state.action {
		case ACTION_SHOOT:
			game.mutex.Unlock()
			game.action_shoot(player)
			game.mutex.Lock()
		case ACTION_COVER:
			action_cover(player)
		case ACTION_RELOAD:
			action_reload(player)
		default:
			fmt.Println("Invalid action:", player.player_state.action)
		}
	}
}

// Method to reset player actions
func (game *Game) reset_player_actions() {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	for player := range game.players {
		player.mutex.Lock()
		player.player_state.action = ""
		player.mutex.Unlock()
	}
}

// Method to handle the shoot action
func (game *Game) action_shoot(player *Player) {
	fmt.Println("Player", player.name, "is shooting...")

	player.mutex.Lock()
	// Find the opponent player
	for opponent := range game.players {
		if opponent != player {
			opponent.mutex.Lock()
			defer opponent.mutex.Unlock()

			if player.player_state.ammo == 0 {
				fmt.Println("Player", player.name, "is out of ammo!")
				break
			} else {
				player.player_state.ammo--
			}

			if opponent.player_state.action != ACTION_COVER {
				opponent.player_state.health--
				fmt.Println("Player", player.name, "shot the opponent", opponent.name)
				break
			}
		}
	}
	player.mutex.Unlock()
}

// Function to handle the reload action
func action_reload(player *Player) {
	fmt.Println("Player", player.name, "is reloading...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.ammo++
	fmt.Println("Player", player.name, "reloaded!")
}

// Function to handle the cover action
func action_cover(player *Player) {
	fmt.Println("Player", player.name, "is covering...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.action = ACTION_COVER
}

// Method to check player health
func (game *Game) check_player_health() {
	game.mutex.Lock()

	var dead_players []*Player

	for player := range game.players {
		player.mutex.Lock()

		if player.player_state.health <= 0 {
			dead_players = append(dead_players, player)
		} else {
			fmt.Printf("Player %s: Health=%d, Ammo=%d\n", player.name, player.player_state.health, player.player_state.ammo)
		}
		player.mutex.Unlock()
	}

	game.mutex.Unlock()

	for _, player := range dead_players {
		fmt.Println("Player", player.name, "is dead...")

		player.mutex.Lock()
		player.player_state.is_dead = true
		player.mutex.Unlock()
	}

	game.check_game_over(dead_players)
}

// Function to get the player's state
func get_player_state(player *Player) {
	fmt.Println("Getting player state for player:", player.name)

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=player_state&status_code=403&message=Player not in game\n"))
		return
	}

	if player.game.game_state != "running" && !strings.HasPrefix(player.game.game_state, "over") {
		fmt.Println("Game is not running.")
		player.conn.Write([]byte("response_type=player_state&status_code=409&message=Game not running\n"))
		return
	}

	if player.player_state.is_dead {
		fmt.Printf("Player %s is dead.\n", player.name)
		player.conn.Write([]byte("response_type=player_state&status_code=200&message=Dead\n"))
	} else {
		fmt.Printf("Player %s: Health=%d, Ammo=%d\n", player.name, player.player_state.health, player.player_state.ammo)
		player.conn.Write([]byte(fmt.Sprintf("response_type=player_state&status_code=200&message=Health=%d, Ammo=%d\n", player.player_state.health, player.player_state.ammo)))
	}
}

// Function to get the opponent's state
func get_opponent_state(player *Player) {
	fmt.Println("Getting opponent state for player:", player.name)

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=opponent_state&status_code=403&message=Player not in game\n"))
		return
	}

	if player.game.game_state != "running" && !strings.HasPrefix(player.game.game_state, "over") {
		fmt.Println("Game is not running.")
		player.conn.Write([]byte("response_type=opponent_state&status_code=409&message=Game not running\n"))
		return
	}

	for opponent := range player.game.players {
		if opponent != player {
			opponent.mutex.Lock()
			defer opponent.mutex.Unlock()

			if opponent.player_state.is_dead {
				fmt.Printf("Opponent %s is dead.\n", opponent.name)
				player.conn.Write([]byte("response_type=opponent_state&status_code=200&message=Dead\n"))
			} else {
				fmt.Printf("Opponent %s: Health=%d\n", opponent.name, opponent.player_state.health)
				player.conn.Write([]byte(fmt.Sprintf("response_type=opponent_state&status_code=200&message=Health=%d\n", opponent.player_state.health)))
			}
		}
	}
}

// Method to check if the game is over
func (game *Game) check_game_over(dead_players []*Player) {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	if len(dead_players) == len(game.players)-1 {
		fmt.Println("Game over: Winner found!")

		for player := range game.players {
			if !player.player_state.is_dead {
				game.game_state = "over:winner:" + player.name
			}
		}
	} else if len(dead_players) == len(game.players) {
		fmt.Println("Game over: Draw!")

		game.game_state = "over:draw"
	} else {
		fmt.Println("Game is still running...")
	}
}

// Function to get the game result
func get_game_result(player *Player) {
	fmt.Println("Getting result for player: ", player.name)

	var game_state, winner_name, player_name string

	// Get the player's game state
	player.mutex.Lock()
	game := player.game
	player_name = player.name
	player.mutex.Unlock()

	// Check if the player is registered in a game
	if game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=game_result&status_code=403&message=Player not in game\n"))
		return
	}

	// Get the game state
	game.mutex.Lock()
	game_state = game.game_state
	game.mutex.Unlock()

	// Check if the game is over
	if !strings.HasPrefix(game_state, "over") {
		fmt.Println("Game is not over.")
		player.conn.Write([]byte("response_type=game_result&status_code=409&message=Game not over\n"))
		return
	}

	// Check if the game is a draw or has a winner
	if strings.HasPrefix(game_state, "over:winner") {
		winner_name = strings.TrimPrefix(game_state, "over:winner:")
	}

	// Send the game result to the player
	if game_state == "over:draw" {
		fmt.Println("Game is a draw.")
		player.conn.Write([]byte("response_type=game_result&status_code=200&message=Draw\n"))
	} else if winner_name != "" {
		if player_name == winner_name {
			fmt.Println("Player", player_name, "won the game.")
			player.conn.Write([]byte("response_type=game_result&status_code=200&message=Win\n"))
		} else {
			fmt.Println("Player", player_name, "lost the game.")
			player.conn.Write([]byte("response_type=game_result&status_code=200&message=Lose\n"))
		}
	}

	game.mutex.Lock()
	// Remove player from the game
	fmt.Println("Removing player", player_name, "from the game.")
	delete(game.players, player)

	// Check if the game has any players left
	if len(game.players) == 0 {
		fmt.Println("No players left in the game. Cleaning up game state.")
		game.game_state = "waiting" // Reset game state for reuse
	}
	game.mutex.Unlock()
}

// Function to get the game state
func get_game_state(player *Player) {
	fmt.Println("Getting game state...")

	player.mutex.Lock()
	game := player.game
	player.mutex.Unlock()

	if game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=game_state&status_code=403&message=Player not in game\n"))
		return
	}

	game.mutex.Lock()
	game_state := game.game_state
	game.mutex.Unlock()

	if game_state == "running" {
		fmt.Println("Game is running.")
		player.conn.Write([]byte("response_type=game_state&status_code=200&message=Running\n"))
	} else if game_state == "reconnect" {
		fmt.Println("Game is waiting for player to reconnect.")
		player.conn.Write([]byte("response_type=game_state&status_code=200&message=Reconnect\n"))
	} else if game_state == "waiting" {
		fmt.Println("Game is waiting.")
		player.conn.Write([]byte("response_type=game_state&status_code=200&message=Waiting\n"))
	} else if game_state == "over:exit" {
		fmt.Println("Game is over due to player exit.")
		player.conn.Write([]byte("response_type=game_state&status_code=200&message=Exit\n"))
	} else {
		fmt.Println("Game is over.")
		player.conn.Write([]byte("response_type=game_state&status_code=200&message=Over\n"))
	}

}

// Function to get the round state
func get_round_state(player *Player) {
	fmt.Println("Getting round state...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=round_state&status_code=403&message=Player not in game\n"))
		return
	}

	if player.game.round_state == "running" {
		fmt.Println("Round is running.")
		player.conn.Write([]byte("response_type=round_state&status_code=200&message=Running\n"))
	} else {
		fmt.Println("Round ended.")
		player.conn.Write([]byte("response_type=round_state&status_code=200&message=End\n"))
	}
}

// Function to reset the game
func reset_game(player *Player) {
	fmt.Println("Resetting the game...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.game == nil {
		fmt.Println("Player is not registered in any game.")
		player.conn.Write([]byte("response_type=reset_game&status_code=403&message=Player not in game\n"))
		return
	}

	fmt.Println("Resetting the game state.")
	player.game.game_state = "waiting" // Reset game state for reuse

	// Reset player state
	player.player_state.health = DEFAULT_HEALTH
	player.player_state.ammo = DEFAULT_AMMO
	player.player_state.action = ""
	player.player_state.is_dead = false

	// Send a confirmation response to the player
	player.conn.Write([]byte("response_type=reset_game&status_code=200&message=Game reset\n"))

	wait_for_players(player)
}

// Function to close the game
func close_game(player *Player) {
	fmt.Println("Exiting the game...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	// Check if the player is registered
	if player.game != nil {

		fmt.Println("Saving player state and removing from the game...")

		// Add player to the disconnected players list
		disconnected_players[player.name] = player

		// Remove player from the game
		player.game.mutex.Lock()
		delete(player.game.players, player)

		fmt.Println("Player removed from the game.")

		// Check if the game has any players left
		if len(player.game.players) == 0 {
			fmt.Println("No players left in the game. Cleaning up game state.")
			player.game.game_state = "waiting" // Reset game state for reuse
			// clear(disconnected_players)
			// clear disconnected players list
			for key := range disconnected_players {
				delete(disconnected_players, key)
			}
		} else {
			fmt.Println("Game waiting for player to reconnect.")
			player.game.game_state = "reconnect"

			// Start the reconnect timer for this game
			go start_reconnect_timer(player.game)
		}
		player.game.mutex.Unlock()
	}

	// Remove the player from the waiting list
	remove_from_waiting_list(player)

	// Send a confirmation response to the player
	fmt.Println("Closing the game for player:", player.name)
	player.conn.Write([]byte("response_type=close_game&status_code=200&message=Goodbye\n"))

	// Close the connection
	fmt.Println("Closing the connection for player:", player.name)
	player.conn.Close()
}

// Function to remove the player from the waiting list
func remove_from_waiting_list(player *Player) {
	fmt.Println("Removing player from the waiting list...")

	// Get the current length of the channel
	channelLength := len(waiting_players)

	for i := 0; i < channelLength; i++ {
		waiting_player := <-waiting_players
		if waiting_player != player {
			waiting_players <- waiting_player
		}
	}
}

// Function to start the reconnect timer
func start_reconnect_timer(game *Game) {
	fmt.Println("Starting reconnect timer...")

	time.Sleep(time.Duration(RECONNECT_TIMEOUT) * time.Second)

	game.mutex.Lock()
	if game.game_state == "reconnect" {
		fmt.Println("Reconnect timer expired. Closing the game...")

		game.game_state = "over:exit"

		// Clear disconnected players list
		// clear(disconnected_players)
		for key := range disconnected_players {
			delete(disconnected_players, key)
		}
	}
	game.mutex.Unlock()
}
