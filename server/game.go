package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var waiting_players = make(chan *Player, MAX_PLAYERS)

func wait_for_players(player *Player) {

	fmt.Println("Waiting for players to join...")

	waiting_players <- player

	if len(waiting_players) == MAX_PLAYERS {

		var game *Game

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

		// Start the game
		game.start_game()
	}
}

func game_ready(player *Player) {

	if player.game != nil {
		if player.game.game_state == "running" {
			fmt.Println("Game is ready.")
			player.conn.Write([]byte("response_type=game_ready&status_code=200&message=Game ready\n"))
		} else if player.game.game_state == "waiting" {
			fmt.Println("Game is in waiting state.")
			player.conn.Write([]byte("response_type=game_ready&status_code=400&message=Game waiting\n"))
		} else {
			fmt.Println("Game is over.")
			player.conn.Write([]byte("response_type=game_ready&status_code=400&message=Game over\n"))
		}

	} else {
		fmt.Println("Game is not ready.")
		player.conn.Write([]byte("response_type=game_ready&status_code=400&message=Game not ready\n"))
	}

}

func create_player(conn net.Conn, player_name string) *Player {

	fmt.Println("Creating player:", player_name)

	player := &Player{
		conn:         conn,
		player_state: PlayerState{health: DEFAULT_HEALTH, ammo: DEFAULT_AMMO, action: ""},
	}
	player.name = player_name
	// conn.Write([]byte(fmt.Sprintf("response_type=join_game&status_code=200&player_name=%s\n", player.name)))
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
		}

		time.Sleep(time.Duration(ROUND_PAUSE_TIME) * time.Second)

	}

	fmt.Println("Game loop ended...")
}

func (game *Game) wait_for_all_actions() bool {

	fmt.Println("Waiting for player actions...")

	timeout := time.After(time.Duration(ACTION_TIMEOUT) * time.Second)

	game.mutex.Lock()
	game.round_state = "running"
	game.mutex.Unlock()

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
			game.mutex.Lock()
			game.round_state = "end"
			game.mutex.Unlock()
			return true
		}

		select {
		case <-timeout:
			fmt.Println("Timeout reached while waiting for player actions.")
			// Send a response to the players
			game.mutex.Lock()
			game.round_state = "timeout"
			game.mutex.Unlock()
			return false
		default:
			fmt.Println("Sleeping for 5 seconds...")
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
}

func set_player_action(player *Player, message string) {

	// action&action_type=shoot
	action := message[len("action&action_type="):]

	fmt.Println("Setting player action:", action)

	// Set the player action
	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.player_state.action = strings.ToUpper(action)

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

func get_player_state(player *Player) {

	fmt.Println("Getting player state for player:", player.name)

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.player_state.is_dead {
		fmt.Printf("Player %s is dead.\n", player.name)
		player.conn.Write([]byte("response_type=player_state&status_code=200&message=Dead\n"))
	} else {
		fmt.Printf("Player %s: Health=%d, Ammo=%d\n", player.name, player.player_state.health, player.player_state.ammo)
		player.conn.Write([]byte(fmt.Sprintf("response_type=player_state&status_code=200&message=Health=%d, Ammo=%d\n", player.player_state.health, player.player_state.ammo)))
	}
}

func get_opponent_state(player *Player) {

	fmt.Println("Getting opponent state for player:", player.name)

	player.mutex.Lock()
	defer player.mutex.Unlock()

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

func (game *Game) check_game_over(dead_players []*Player) {

	// fmt.Println("Checking game over...")

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

func get_game_result(player *Player) {

	fmt.Println("Getting result for player: ", player.name)

	var game_state, winner_name, player_name string

	player.mutex.Lock()
	game := player.game
	player_name = player.name
	player.mutex.Unlock()

	game.mutex.Lock()
	fmt.Println("Got lock on game.mutex...")
	game_state = game.game_state
	if strings.HasPrefix(game_state, "over:winner") {
		winner_name = strings.TrimPrefix(game_state, "over:winner:")
	}
	game.mutex.Unlock()

	if !strings.HasPrefix(game_state, "over") {
		fmt.Println("Game is not over.")
		player.conn.Write([]byte("response_type=game_result&status_code=400&message=Game not over\n"))
		return
	}

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

func get_round_state(player *Player) {

	fmt.Println("Getting round state...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.game.round_state == "running" {
		fmt.Println("Round is running.")
		player.conn.Write([]byte("response_type=round_state&status_code=200&message=Running\n"))
	} else if player.game.round_state == "timeout" {
		fmt.Println("Round timed out.")
		player.conn.Write([]byte("response_type=round_state&status_code=200&message=Timeout\n"))
	} else {
		fmt.Println("Round ended.")
		player.conn.Write([]byte("response_type=round_state&status_code=200&message=End\n"))
	}

}

func reset_game(player *Player) {

	fmt.Println("Resetting the game...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	// Check if the player is registered
	if player.game != nil {
		// Remove player from the game
		player.game.mutex.Lock()
		delete(player.game.players, player)
		player.game.mutex.Unlock()

		// Check if the game has any players left
		if len(player.game.players) == 0 {
			player.game.game_state = "waiting" // Reset game state for reuse
		}

		player.game = nil
	}

	// Reset player state
	player.player_state.health = DEFAULT_HEALTH
	player.player_state.ammo = DEFAULT_AMMO
	player.player_state.action = ""
	player.player_state.is_dead = false

	// Send a confirmation response to the player
	player.conn.Write([]byte("response_type=reset_game&status_code=200&message=Game reset\n"))

	wait_for_players(player)

}

func close_game(player *Player) {

	fmt.Println("Exiting the game...")

	player.mutex.Lock()
	defer player.mutex.Unlock()

	// Check if the player is registered
	if player.game != nil {
		// Remove player from the game
		player.game.mutex.Lock()
		delete(player.game.players, player)
		player.game.mutex.Unlock()

		// Check if the game has any players left
		if len(player.game.players) == 0 {
			player.game.game_state = "waiting" // Reset game state for reuse
		}

		player.game = nil
	}

	// Send a confirmation response to the player
	player.conn.Write([]byte("response_type=close_game&status_code=200&message=Goodbye\n"))

	// Close the connection
	player.conn.Close()

}
