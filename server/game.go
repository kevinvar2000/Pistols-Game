package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var waiting_players = make(chan *Player, MAX_PLAYERS)

func wait_for_players(player *Player) {

	fmt.Println("Waiting for players to join...")

	waiting_players <- player

	if len(waiting_players) == MAX_PLAYERS {
		// Create a new game
		game := create_game()

		// Add players to the game
		for i := 0; i < MAX_PLAYERS; i++ {
			player := <-waiting_players
			game.players[player] = true
			fmt.Println("Player joined the game:", player.name)
		}

		broadcast_message(game, "response_type=game_ready&status_code=200&message=Game ready\n")

		// Start the game
		game.start_game()
	}
}

func create_player(conn net.Conn, player_name string) *Player {

	fmt.Println("Creating player:", player_name)

	player := &Player{
		conn:         conn,
		player_state: PlayerState{health: DEFAULT_HEALTH, ammo: DEFAULT_AMMO, action: ""},
	}
	player.name = player_name
	conn.Write([]byte(fmt.Sprintf("response_type=join_game&status_code=200&player_name=%s\n", player.name)))
	return player
}

func create_game() *Game {

	fmt.Println("Creating a new game...")

	game := &Game{
		players:    make(map[*Player]bool),
		game_state: "waiting",
		mutex:      &sync.Mutex{},
	}
	return game
}

func (game *Game) start_game() {

	fmt.Println("Starting the game...")

	game.game_state = "running"

	broadcast_message(game, "response_type=start_game&status_code=200&message=Game started\n")

	// Start the game loop
	go game.game_loop()

}

func (game *Game) game_loop() {

	fmt.Println("Game loop started...")

	for {
		// Check if the game is running
		if game.game_state != "running" {
			fmt.Println("Game state is not running...")
			break
		}

		if game.wait_for_all_actions() {

			// Check player actions
			game.check_player_actions()

			// Reset player actions
			game.reset_player_actions()

			// Check player health
			game.check_player_health()

			// Check game state
			game.check_game_state()

		}

		time.Sleep(time.Duration(ROUND_PAUSE_TIME) * time.Second)

	}

	fmt.Println("Game loop ended...")
}

func (game *Game) wait_for_all_actions() bool {

	fmt.Println("Waiting for player actions...")

	timeout := time.After(time.Duration(ACTION_TIMEOUT) * time.Second)

	for {
		all_ready := true
		game.mutex.Lock()
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
			return true
		}

		select {
		case <-timeout:
			fmt.Println("Timeout reached while waiting for player actions.")
			// Send a response to the players
			broadcast_message(game, "response_type=timeout&status_code=200&message=Timeout reached\n")
			return false
		default:
			fmt.Println("Sleeping for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
}

func set_player_action(player *Player, action string) {

	fmt.Println("Setting player action:", action)

	// Set the player action
	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.action = action

}

func (game *Game) check_player_actions() {

	// fmt.Println("Checking player actions...")

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

func (game *Game) reset_player_actions() {

	// fmt.Println("Resetting player actions...")
	game.mutex.Lock()
	defer game.mutex.Unlock()

	for player := range game.players {
		player.mutex.Lock()
		player.player_state.action = ""
		player.mutex.Unlock()
	}

}

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

func action_reload(player *Player) {

	fmt.Println("Player", player.name, "is reloading...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.ammo++
	fmt.Println("Player", player.name, "reloaded!")

}

func action_cover(player *Player) {
	fmt.Println("Player", player.name, "is covering...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.action = ACTION_COVER
}

func (game *Game) check_player_health() {

	// fmt.Println("Checking player health...")

	game.mutex.Lock()
	defer game.mutex.Unlock()

	for player := range game.players {

		player.mutex.Lock()

		if player.player_state.health <= 0 {
			// Player is dead
			game.player_dead(player)
		} else {
			fmt.Printf("Player %s: Health=%d, Ammo=%d\n", player.name, player.player_state.health, player.player_state.ammo)
			player.conn.Write([]byte(fmt.Sprintf("response_type=player_health&status_code=200&message=Health=%d, Ammo=%d\n", player.player_state.health, player.player_state.ammo)))
		}
		player.mutex.Unlock()
	}

}

func (game *Game) player_dead(player *Player) {

	fmt.Println("Player", player.name, "is dead...")

	game.game_state = "over"

	// Send a response to the player
	player.conn.Write([]byte("response_type=player_dead&status_code=200&message=You are dead\n"))

}

func (game *Game) check_game_state() {

	// fmt.Println("Checking game state...")
	// TODO: Implement game over condition

	game.mutex.Lock()
	defer game.mutex.Unlock()

	if game.game_state == "over" {
		game.game_over()
	}

}

func (game *Game) game_over() {

	fmt.Println("Game over...")

	// Find the winner
	var winner *Player
	for player := range game.players {
		winner = player
		break
	}

	// Check if there is a winner
	if winner == nil {
		// If no players are left, log this scenario
		fmt.Println("No players left alive. No winner!")
		broadcast_message(game, "response_type=game_over&status_code=200&message=No winner\n")
		return
	} else {
		// Send a response to the winner
		fmt.Println("Winner:", winner.name)
		winner.conn.Write([]byte("response_type=game_over&status_code=200&message=You are the winner\n"))
	}

	// Close the connections
	// TODO: Implement a better way to close connections
	// ? Should we close the connections here?
	for player := range game.players {
		fmt.Println("Closing connection for player:", player.name)
		player.conn.Close()
	}

	fmt.Println("Game over. Closing the game...")

	// Clear the players
	for player := range game.players {
		delete(game.players, player)
	}

	fmt.Println("Game players:", game.players)

	// Reset the game state
	game.game_state = "waiting"

}

func exit_game(player *Player) {

	fmt.Println("Exiting the game...")

	// Send a confirmation response to the player
	player.conn.Write([]byte("response_type=exit&status_code=200&message=Goodbye\n"))

	// Close the connection
	player.conn.Close()

}
