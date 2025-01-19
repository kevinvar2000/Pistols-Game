import time
import tkinter as tk
from tkinter import messagebox
from const import WIN_SIZE, BUTTON_FONT, LABEL_FONT, REQUEST_INTERVAL, ELEMENT_SIZE, WIN_BG, BTN_BG, BTN_FG, LABEL_BG, LABEL_FG, ERROR_BG, ERROR_FG, INACTIVE_TIMEOUT, CHECK_INTERVAL, DEFAULT_HEALTH, DEFAULT_AMMO

class GameWindow:
    def __init__(self, root, client, client_name):
        self.client = client
        self.client_name = client_name
        self.check_activity = True
        self.cancel_call = False
        self.after_reconnect = False
        self.game_over = False

        self.root = root
        self.root.title("Game")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)
        self.root.configure(bg=WIN_BG)

        # Start the waiting room until game is ready
        self.waiting_room()

    def send_action(self, action):

        self.check_activity = False

        # Disable actions while waiting for response
        self.disable_actions()

        print(f"*** Sending action: {action} ***")

        # Send the action request to the server
        response = self.client.send_request(action, "action")
        print(f"Response in send_action: {response}")

        if response.get("response_type") == "action":
            if response.get("status_code") == "200":
                self.update_game(response)
            elif response.get("status_code") == "400":
                print(f"Invalid action: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.enable_actions)
            elif response.get("status_code") == "403":
                print("Player not in game. Returning to connect window...")
                self.return_to_connect_window()
            elif response.get("status_code") == "409":
                print("Game is not running yet or action is already set. Checking again shortly...")
                self.root.after(REQUEST_INTERVAL, self.send_action, action)
            else:
                print(f"Failed to send action: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, lambda: self.send_action(action))
        elif response.get("status") == "error":
            print(f"Response error: {response.get('message')}")
            self.update_labels(error_message=response.get("message"))
            self.root.after(REQUEST_INTERVAL, self.return_to_connect_window)

    def shoot(self):
        self.send_action("request_type=action&action_type=shoot")

    def reload(self):
        self.send_action("request_type=action&action_type=reload")

    def cover(self):
        self.send_action("request_type=action&action_type=cover")

    def disable_actions(self, reconnect=False):
        print("*** Disabling actions ***")

        # Update the info label
        if reconnect:
            self.info_label.config(text="Opponent is reconnecting...")
        else:
            self.info_label.config(text="Waiting for response...")

        # Disable all widgets
        for widget in self.root.winfo_children():
            widget.config(state=tk.DISABLED)

        self.root.update_idletasks()

    def enable_actions(self):
        print("*** Enabling actions ***")

        # Update the info label
        self.info_label.config(text="Choose an action:")

        # Enable all widgets
        for widget in self.root.winfo_children():
            widget.config(state=tk.NORMAL)

        self.root.update_idletasks()

    def waiting_room(self):
        print("*** Waiting room ***")

        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Display a waiting message
        waiting_label = tk.Label(self.root, text="Waiting for other player to join...", font=LABEL_FONT)
        waiting_label.pack(pady=50)

        self.error_label = tk.Label(self.root, text="", font=LABEL_FONT, fg=ERROR_FG, bg=ERROR_BG, relief="solid", wraplength=400, justify="center")

        self.root.update_idletasks()

        # Check if the game is ready
        self.root.after(REQUEST_INTERVAL, self.start_game)

    def start_game(self):
        print("*** Checking if game is ready ***")

        # Send a request to check if the game is ready
        response = self.client.send_request("request_type=start_game", "start_game")
        print(f"Response in start_game: {response}")

        if response.get("response_type") == "start_game":
            if response.get("status_code") == "200":
                
                # Game is ready, load the game window
                print("Game is ready! Starting the game...")
                self.load_game()

                if response.get("player_state"):
                    self.update_player_state(response.get("player_state"))
                if response.get("opponent_state"):
                    self.update_opponent_state(response.get("opponent_state"))
                if response.get("game_state"):
                    self.update_game_state(response.get("game_state"))

            else:
                print(f"Game not ready: {response.get('message')}")
                self.root.after(REQUEST_INTERVAL, self.start_game)
        elif response.get("status") == "error":
            print(f"Response error: {response.get('message')}")
            self.update_labels(error_message=response.get("message"))
            self.root.after(REQUEST_INTERVAL, self.return_to_connect_window)


    def update_game(self, response):
        print("*** Updating game ***")

        # Parsed response: {'response_type': 'action', 'status_code': '200', 'player_state': 'health=3,ammo=0', 'opponent_state': 'health=2', 'game_state': 'running'\n'}

        if response.get("response_type") == "action" and response.get("status_code") == "200":

            if response.get("player_state"):
                self.update_player_state(response.get("player_state"))
            if response.get("opponent_state"):
                self.update_opponent_state(response.get("opponent_state"))
            if response.get("game_state"):
                self.update_game_state(response.get("game_state"))

            if self.game_over:
                print("Game is over. Ending game...")
                return
            # Enable actions after updating the game
            self.root.after(REQUEST_INTERVAL, self.enable_actions)

            self.check_activity = True
            self.last_action_time = time.time()

        elif response.get("status") == "error":
            print(f"Response error: {response.get('message')}")
            self.update_labels(error_message=response.get("message"))
            self.root.after(REQUEST_INTERVAL, self.return_to_connect_window)

    def load_game(self):
        print("*** Loading game window ***")

        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Client name display
        self.name_label = tk.Label(self.root, text=f"Name: {self.client_name}", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.name_label.pack(pady=10)

        # Health and Ammo display
        self.health_label = tk.Label(self.root, text=f"Health:{DEFAULT_HEALTH}", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.health_label.pack(pady=10)

        self.ammo_label = tk.Label(self.root, text=f"Ammo:{DEFAULT_AMMO}", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.ammo_label.pack(pady=10)

        # Opponent's Health display
        self.opponent_health_label = tk.Label(self.root, text=f"Opponent Health:{DEFAULT_HEALTH}", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.opponent_health_label.pack(pady=5)
    
        # Info and Error labels
        self.info_label = tk.Label(self.root, text="Choose an action:", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.info_label.pack(pady=10)

        self.error_label = tk.Label(self.root, text="", font=LABEL_FONT, fg=ERROR_FG, bg=ERROR_BG, relief="solid", wraplength=400, justify="center")

        # Buttons for actions
        tk.Button(self.root, text="Shoot", font=BUTTON_FONT, command=self.shoot, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)
        tk.Button(self.root, text="Reload", font=BUTTON_FONT, command=self.reload, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)
        tk.Button(self.root, text="Cover", font=BUTTON_FONT, command=self.cover, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)

        # Handle window close event
        self.root.protocol("WM_DELETE_WINDOW", self.close_window)

        print("Starting the game...")

        # Set flags
        self.check_activity = True
        self.cancel_call = False
        self.game_over = False

        # Start checking for inactivity
        self.last_action_time = time.time()

        # Check for inactivity every 5 seconds
        self.root.after(CHECK_INTERVAL, self.check_inactivity)

        # Check the game state
        self.root.after(CHECK_INTERVAL, self.check_game_state)

    def check_inactivity(self):

        if not self.check_activity:
            self.root.after(CHECK_INTERVAL, self.check_inactivity)
            return

        current_time = time.time()

        if current_time - self.last_action_time > INACTIVE_TIMEOUT:
            print(f"Player has been inactive for {INACTIVE_TIMEOUT} seconds. Sending a reload action...")
            self.update_labels(error_message=f"You have been inactive for {INACTIVE_TIMEOUT} seconds. Sending a reload action...")
            self.reload()
            self.root.after(REQUEST_INTERVAL, lambda: self.update_labels(error_message=""))

        self.root.after(CHECK_INTERVAL, self.check_inactivity)

    def update_player_state(self, player_state):
        print("*** Getting player state ***")

        if player_state == "dead":
            print("Player is dead. Ending game...")

            # Update the labels
            self.update_labels(health=0, ammo=0, info_message="You are dead. Game over.")
        else:
            try:
                state_pairs = dict(pair.split("=") for pair in player_state.split(","))
                self.health = int(state_pairs.get("health", 0))
                self.ammo = int(state_pairs.get("ammo", 0))
                print(f"Updated player state: Health={self.health}, Ammo={self.ammo}")

                self.update_labels(self.health, self.ammo)
            except (ValueError, KeyError) as e:
                print(f"Error parsing player state: {e}")
                self.update_labels(error_message="Error parsing player state")

    def update_opponent_state(self, opponent_state):
        print("*** Getting opponent state ***")

        if opponent_state == "dead":
            print("Opponent is dead. Ending game...")

            # Update the labels
            self.update_labels(opponent_health=0, info_message="Opponent is dead. You win!")
        elif "health" in opponent_state:

            try:
                opponent_health = int(opponent_state.split("=")[1].strip())
                print(f"Updated opponent state: Health={opponent_health}")
            except ValueError as e:
                print(f"Error parsing opponent state: {e}")
                self.update_labels(error_message="Error parsing opponent state")

            # Update opponent state
            self.update_labels(opponent_health=opponent_health)

    def check_game_state(self):

        if self.cancel_call:
            return

        print("*** Checking game state ***")


        # Send a request to check the game state
        response = self.client.send_request("request_type=game_state", "game_state")
        print(f"Response in check_game_state: {response}")

        if response.get("response_type") == "game_state":
            if response.get("status_code") == "200":
                message_state = response.get("message").strip()
                print(f"Message state: {message_state}")

                game_state = message_state.split()[1].strip()

                self.update_game_state(game_state)
            else:
                print(f"Failed to get game state: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(CHECK_INTERVAL, self.check_game_state)
        elif response.get("status") == "error":
            print(f"Response error: {response.get('message')}")
            self.update_labels(error_message=response.get("message"))
            self.root.after(REQUEST_INTERVAL, self.return_to_connect_window)


    def update_game_state(self, game_state):
        print("*** Getting game state ***")

        print(f"Game state: {game_state}")

        if game_state == "running":
            print("Game is running. Checking game state...")
            self.root.after(REQUEST_INTERVAL, self.check_game_state)

            if self.after_reconnect:
                self.enable_actions()
                self.after_reconnect = False
                self.check_activity = True

        elif "over:" in game_state:
            print("Game is over. Resolving game result...")
            self.resolve_result(game_state)
        elif game_state == "over":
            print("Game is over. Checking results...")
            self.update_labels(info_message="Game over. Checking results...")
        elif game_state == "waiting":
            print("Game is waiting for players. Returning to the waiting room...")
            self.waiting_room()
        elif game_state == "reconnect":
            print("Opponent is reconnecting. Disabling actions...")
            self.disable_actions(reconnect=True)
            self.after_reconnect = True
            self.check_activity = False
            self.root.after(REQUEST_INTERVAL, self.check_game_state)
        else:
            print(f"Invalid game state: {game_state}")
            self.update_labels(error_message="Invalid game state")
            self.root.after(REQUEST_INTERVAL, self.check_game_state)


    def resolve_result(self, game_state):
        print("*** Resolving game result ***")

        game_result = "Game over: "

        if "winner" in game_state:
            winner = game_state.split(":")[2].strip()
            if winner == self.client_name:
                game_result += "You win!"
            else:
                game_result += "You lose!"

        elif "draw" in game_state:
            game_result += "It's a draw!"

        print(game_result)

        self.game_over = True

        # Update the labels
        self.root.after(0, lambda: self.update_labels(info_message=game_result))

        # Ask player to play again or quit
        self.root.after(REQUEST_INTERVAL, lambda: self.ask_play_again(game_result))


    def ask_play_again(self, result_message):

        self.check_activity = False
        self.cancel_call = True

        # Cancel all pending callbacks
        self.cancel_callbacks()

        # Ask the player if they want to play again
        result = messagebox.askquestion("Game Over", f"{result_message}\nDo you want to play again?", icon='question')

        if result == 'yes':
            print("Player chose to play again. Returning to the waiting room...")
            self.reset_game()
        else:
            print("Player chose to quit. Closing the game...")
            self.close_window()

    def reset_game(self):
        print("*** Resetting game ***")

        # Send a request to reset the game
        response = self.client.send_request("request_type=reset_game", "reset_game")
        print(f"Response in reset_game: {response}")

        if response.get("response_type") == "reset_game":
            if response.get("status_code") == "200":
                print("Game reset successfully.")
                self.waiting_room()
            else:
                print(f"Failed to reset game: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.reset_game)
        elif response.get("status") == "error":
            print(f"Response error: {response.get('message')}")
            self.update_labels(error_message=response.get("message"))
            self.root.after(REQUEST_INTERVAL, self.return_to_connect_window)

    def update_labels(self, health=None, ammo=None, opponent_health=None, info_message=None, error_message=None):
        print("*** Updating labels ***")

        if health is not None:
            print(f"Updating health: {health}")
            self.health_label.config(text=f"Health: {health}")
        if ammo is not None:
            print(f"Updating ammo: {ammo}")
            self.ammo_label.config(text=f"Ammo: {ammo}")
        if opponent_health is not None:
            print(f"Updating opponent health: {opponent_health}")
            self.opponent_health_label.config(text=f"Opponent Health: {opponent_health}")
        if info_message is not None:
            print(f"Updating info label: {info_message}")
            self.info_label.config(text=info_message)
        if error_message is not None:
            print(f"Updating error label: {error_message}")
            self.error_label.config(text=error_message)
            if error_message != "":
                self.error_label.pack(pady=10)
            else:    
                self.error_label.pack_forget()
        
        self.root.update()
        self.root.update_idletasks()

    def cancel_callbacks(self):
        print("*** Cancelling callbacks ***")

        try:
            self.root.after_cancel(self.check_inactivity)
            self.root.after_cancel(self.check_game_state)
        except Exception as e:
            print(f"Error cancelling callbacks: {e}")
        else:
            print("Callbacks cancelled successfully.")

    def close_window(self):
        print("*** Closing game window ***")

        # Cancel all pending callbacks
        self.cancel_callbacks()

        # Send a request to close the game
        self.client.close_game()

        # Destroy the root window
        self.root.destroy()

    
    def return_to_connect_window(self):
        print("*** Returning to connect window ***")

        # Cancel all pending callbacks
        self.cancel_callbacks()

        # Destroy the root window
        self.root.destroy()

        # Return to the connect window
        from connect_win import ConnectWindow
        ConnectWindow()