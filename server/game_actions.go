package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

func (s *GameServer) Start_game(client *Client) error {

	fmt.Println("***Starting the game***")

	client.conn.Write([]byte("STARTING_GAME\n"))
	client.room.Broadcast(fmt.Sprintf("%s has joined. Game starts now!", client.name), client)

	fmt.Println("***Game started***")
	fmt.Println()
	return nil
}

func (s *GameServer) Handle_game_actions(client *Client) {

	fmt.Println("***Handling player actions during the game***")

	reader := bufio.NewReader(client.conn)
	client.conn.Write([]byte("GAME: Enter your action (SHOOT, COVER, RELOAD) or type DISCONNECT to leave: \n"))

	for {
		// Read player action
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection:", client.name)
			}
			fmt.Printf("Error reading message from client %s: %v\n", client.name, err)
			client.room.Remove_client(client, s)
			break
		}

		message = strings.TrimSpace(message)
		fmt.Println("Received message from client:", message)

		message = strings.ToUpper(message)

		// Check for correct message formatd
		if message != "SHOOT" && message != "COVER" && message != "RELOAD" && message != "DISCONNECT" {
			client.conn.Write([]byte("GAME: Invalid action. Enter SHOOT, COVER, RELOAD, or DISCONNECT: \n"))
			continue
		}

		// Handle client disconnection
		if message == "DISCONNECT" {
			client.room.Remove_client(client, s)
			break
		}

		// **Handle player actions**
		// if strings.HasPrefix(message, "PLAYER_ACTION:") {
		// 	client.room.handleAction(client, strings.TrimPrefix(message, "PLAYER_ACTION:"))
		// }
		client.room.Handle_action(client, message)

		client.conn.Write([]byte("GAME: Enter your action (SHOOT, COVER, RELOAD) or type DISCONNECT to leave: \n"))

	}

	fmt.Println("***After handling player actions during the game***")
	fmt.Println()
}

func (r *Room) Handle_action(client *Client, action string) {

	fmt.Println("***Handling player action in the room***")

	r.mu.Lock()
	state := r.states[client]
	state.action = strings.ToUpper(action)
	fmt.Println("Client:", client.name, ", Action:", state.action)

	fmt.Println("Current actions in the room:")
	for _, client := range r.clients {
		fmt.Println(client.name, ":", r.states[client].action)
	}
	r.mu.Unlock() // Unlock immediately after updating state

	// Trigger waiting logic
	go r.Wait_for_action()

	fmt.Println("***After handling player action in the room***")
	fmt.Println()
}

func (r *Room) Wait_for_action() {

	fmt.Println("***Waiting for second action***")

	// Timeout duration (e.g., 10 seconds)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Println("Timeout reached while waiting for second player's action.")

			fmt.Println("Aquiring lock in wait for action timeout.")
			r.mu.Lock()
			fmt.Println("Lock acquired in wait for action timeout.")

			// Apply default action for any missing responses
			for _, client := range r.clients {
				if r.states[client].action == "" {
					r.states[client].action = "COVER" // Default action
					fmt.Printf("Default action applied for %s: COVER\n", client.name)
				}
			}

			player1_action := r.states[r.clients[0]]
			player2_action := r.states[r.clients[1]]

			r.mu.Unlock() // Unlock before resolving

			r.Resolve_action(player1_action, player2_action)
			r.Reset_action()

			fmt.Println("Exiting after timeout.")
			return

		case <-ticker.C:

			fmt.Println("Acquiring lock in wait for action.")
			r.mu.Lock()
			fmt.Println("Lock acquired in wait for action.")
			allReceived := r.states[r.clients[0]].action != "" && r.states[r.clients[1]].action != ""
			r.mu.Unlock()

			if allReceived {
				fmt.Println("Both actions received before timeout.")

				player1_action := r.states[r.clients[0]]
				player2_action := r.states[r.clients[1]]

				r.Resolve_action(player1_action, player2_action)
				r.Reset_action()
				fmt.Println("Exiting after both actions received.")
				return
			}
		}
	}
}

func (r *Room) Resolve_action(p1, p2 *PlayerState) {

	fmt.Println("***Resolving player actions in the room***")

	fmt.Println("Acquiring lock in resolve action.")
	r.mu.Lock()
	fmt.Println("Lock acquired in resolve action.")

	fmt.Println("Resolving actions:", p1.action, p2.action)
	if p1.action == "SHOOT" && p2.action != "COVER" {
		fmt.Println("Player 1 shoots Player 2.")
		p2.health -= 1
	}
	if p2.action == "SHOOT" && p1.action != "COVER" {
		fmt.Println("Player 2 shoots Player 1.")
		p1.health -= 1
	}

	if p1.action == "RELOAD" {
		fmt.Println("Player 1 reloads.")
		p1.ammo += 1
	}

	if p2.action == "RELOAD" {
		fmt.Println("Player 2 reloads.")
		p2.ammo += 1
	}

	var result_msg string
	if p1.health <= 0 && p2.health <= 0 {
		fmt.Println("Both players are dead.")
		result_msg = "RESULT: Both Players are Dead!"
	} else if p1.health <= 0 {
		fmt.Println("Player 2 wins.")
		result_msg = "RESULT: Player 2 Wins!"
	} else if p2.health <= 0 {
		fmt.Println("Player 1 wins.")
		result_msg = "RESULT: Player 1 Wins!"
	} else {
		fmt.Println("Players are still alive.")
		result_msg = fmt.Sprintf("RESULT: Player1 - %d HP, Player2 - %d HP", p1.health, p2.health)
	}

	r.mu.Unlock()

	r.Broadcast(result_msg, nil)
	fmt.Println("***After resolving player actions in the room***")
	fmt.Println()
}

func (r *Room) Reset_action() {

	fmt.Println("***Resetting actions for all clients in the room***")

	for _, client := range r.clients {
		r.states[client].action = ""
	}

	fmt.Println("***After resetting actions for all clients***")
	fmt.Println()
}

func (s *GameServer) Wait_for_player(client *Client) error {

	fmt.Println("***Waiting for second player to join***")

	client.conn.Write([]byte("WAITING_FOR_PLAYER\n"))
	fmt.Println("Waiting for another player to join the room:", client.room.clients[0].name)

	timeout := time.After(30 * time.Second) // Set a 30-second timeout
	fmt.Println("Timeout set to 30 seconds")

	for {
		if client.room == nil {
			return fmt.Errorf("client.room is nil")
		}

		client.room.mu.Lock()
		currentCount := len(client.room.clients)
		// fmt.Printf("Current room count: %d\n", currentCount)

		if currentCount == 2 {
			client.room.mu.Unlock()
			fmt.Println("Both players have joined. Proceeding with the game.")
			break
		}
		client.room.mu.Unlock()

		select {
		case <-timeout:
			fmt.Println("Timeout reached: No second player joined.")
			client.conn.Write([]byte("TIMEOUT: No second player joined.\n"))
			client.room.Remove_client(client, s)
			return fmt.Errorf("timeout reached: no second player joined")
		default:
			time.Sleep(100 * time.Millisecond) // Prevent busy looping
		}
	}

	fmt.Println("***After waiting for second player to join***")
	fmt.Println()
	return nil
}
