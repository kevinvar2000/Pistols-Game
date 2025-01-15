import time
import tkinter as tk
from tkinter import messagebox
from const import WIN_SIZE, BUTTON_FONT, LABEL_FONT, REQUEST_INTERVAL, ELEMENT_SIZE, WIN_BG, BTN_BG, BTN_FG, LABEL_BG, LABEL_FG, ERROR_BG, ERROR_FG, INACTIVE_TIMEOUT, CHECK_INTERVAL

class GameWindow:
    def __init__(self, root, client, client_name):
        self.client = client
        self.client_name = client_name
        self.check_activity = True
        self.check_result = False
        self.cancel_call = False
        self.after_reconnect = False

        self.root = root
        self.root.title("Game")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)
        self.root.configure(bg=WIN_BG)

        # Start the waiting room until game is ready
        self.waiting_room()

    def send_action(self, action):
        # Update the last action time
        self.last_action_time = time.time()

        # Disable actions while waiting for response
        self.disable_actions()

        print(f"*** Sending action: {action} ***")

        # Send the action request to the server
        response = self.client.send_request(action, "action")
        print(f"Response in send_action: {response}")

        if response.get("response_type") == "action":
            if response.get("status_code") == "200":
                # Check the round state if action was successful
                self.check_round_state()
            else:
                print(f"Failed to send action: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, lambda: self.send_action(action))
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, lambda: self.send_action(action))

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
            self.info_label.config(text="Reconnecting... Please wait.")
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
        self.root.update_idletasks()

        # Check if the game is ready
        self.check_game_ready()

    def check_game_ready(self):
        print("*** Checking if game is ready ***")

        # Send a request to check if the game is ready
        response = self.client.send_request("request_type=game_ready", "game_ready")
        print(f"Response in check_game_ready: {response}")

        if response.get("response_type") == "game_ready":
            if response.get("status_code") == "200":
                # Game is ready, load the game window
                print("Game is ready! Starting the game...")
                self.load_game()
            else:
                print(f"Game not ready: {response.get('message')}")
                self.root.after(REQUEST_INTERVAL, self.check_game_ready)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.check_game_ready)

    def check_round_state(self):
        print("*** Checking round state ***")

        # Send a request to check the round state
        response = self.client.send_request("request_type=round_state", "round_state")
        print(f"Response in check_round_state: {response}")

        if response.get("response_type") == "round_state":
            if response.get("status_code") == "200":
                state_message = response.get("message").strip()
                print(f"Round state: {state_message}")

                if state_message == "Running":
                    print("Round is still running. Checking again shortly...")
                    self.root.after(REQUEST_INTERVAL, self.check_round_state)
                elif state_message == "End":
                    print("Round ended. Processing results...")
                    self.get_player_state()
                    self.root.after(2000, self.enable_actions)
            else:
                print(f"Failed to get round state: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.check_round_state)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.check_round_state)

    def load_game(self):
        print("*** Loading game window ***")

        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Client name display
        self.name_label = tk.Label(self.root, text=f"Name: {self.client_name}", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.name_label.pack(pady=10)

        # Health and Ammo display
        self.health_label = tk.Label(self.root, text="Health: Loading...", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.health_label.pack(pady=10)

        self.ammo_label = tk.Label(self.root, text="Ammo: Loading...", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.ammo_label.pack(pady=10)

        # Opponent's Health display
        self.opponent_health_label = tk.Label(self.root, text="Opponent Health: Loading...", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.opponent_health_label.pack(pady=5)
    
        # Info and Error labels
        self.info_label = tk.Label(self.root, text="Choose an action:", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG)
        self.info_label.pack(pady=10)

        self.error_label = tk.Label(self.root, text="", font=LABEL_FONT, fg=ERROR_FG, bg=ERROR_BG, relief="solid")

        # Buttons for actions
        tk.Button(self.root, text="Shoot", font=BUTTON_FONT, command=self.shoot, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)
        tk.Button(self.root, text="Reload", font=BUTTON_FONT, command=self.reload, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)
        tk.Button(self.root, text="Cover", font=BUTTON_FONT, command=self.cover, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=5)

        # Handle window close event
        self.root.protocol("WM_DELETE_WINDOW", self.close_window)

        print("Starting the game...")

        # Set flags
        self.check_activity = True
        self.check_result = False
        self.cancel_call = False

        # Start fetching the player state asynchronously
        self.root.after(0, self.get_player_state)

        # Start checking for inactivity
        self.last_action_time = time.time()

        # Check for inactivity every 5 seconds
        self.root.after(CHECK_INTERVAL, self.check_inactivity)

        # Check the game state
        self.root.after(CHECK_INTERVAL, self.check_game_state)

    def check_inactivity(self):

        if not self.check_activity:
            return

        current_time = time.time()

        if current_time - self.last_action_time > INACTIVE_TIMEOUT:
            print(f"Player has been inactive for {INACTIVE_TIMEOUT} seconds. Sending a reload action...")
            self.update_labels(error_message=f"You have been inactive for {INACTIVE_TIMEOUT} seconds. Sending a reload action...")
            self.reload()
            self.root.after(REQUEST_INTERVAL, lambda: self.update_labels(error_message=""))

        self.root.after(CHECK_INTERVAL, self.check_inactivity)

    def get_player_state(self):
        print("*** Getting player state ***")

        # Send a request to get the player state
        response = self.client.send_request("request_type=player_state", "player_state")
        print(f"Response in get_player_state: {response}")

        if response.get("response_type") == "player_state":
            if response.get("status_code") == "200":
                state_message = response.get("message").strip()
                print(f"Player state: {state_message}")

                if state_message == "Dead":
                    print("Player is dead. Ending game...")

                    # Update the labels
                    self.update_labels(health=0, ammo=0, info_message="You are dead. Game over.")
                else:
                    try:
                        state_pairs = dict(pair.split("=") for pair in state_message.split(", "))
                        self.health = int(state_pairs.get("Health", 0))
                        self.ammo = int(state_pairs.get("Ammo", 0))
                        print(f"Updated player state: Health={self.health}, Ammo={self.ammo}")

                        self.update_labels(self.health, self.ammo)
                    except (ValueError, KeyError) as e:
                        print(f"Error parsing player state: {e}")
                        self.update_labels(error_message="Error parsing player state")
                        self.root.after(REQUEST_INTERVAL, self.get_player_state)
                        return
            else:
                print(f"Failed to get player state: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.get_player_state)
                return
        elif response.get("response_type") == "game_ready":
            print("Received 'game_ready' during player state check. Retrying...")
            self.root.after(REQUEST_INTERVAL, self.get_player_state)
            return
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.get_player_state)
            return        

        # Check the game result
        self.check_game_result()

    def get_opponent_state(self):
        print("*** Getting opponent state ***")

        # Send a request to get the opponent state
        response = self.client.send_request("request_type=opponent_state", "opponent_state")
        print(f"Response in get_opponent_state: {response}")

        if response.get("response_type") == "opponent_state":
            if response.get("status_code") == "200":
                state_message = response.get("message").strip()
                print(f"Opponent state: {state_message}")

                if state_message == "Dead":
                    print("Opponent is dead. Ending game...")

                    # Update the labels
                    self.update_labels(opponent_health=0, info_message="Opponent is dead. You win!")
                else:
                    try:
                        state_pairs = dict(pair.split("=") for pair in state_message.split(", "))
                        opponent_health = int(state_pairs.get("Health", 0))
                        print(f"Updated opponent state: Health={opponent_health}")

                        # Update opponent state
                        self.update_labels(opponent_health=opponent_health)
                    except (ValueError, KeyError) as e:
                        print(f"Error parsing opponent state: {e}")
                        self.update_labels(error_message="Error parsing opponent state")
                        self.root.after(REQUEST_INTERVAL, self.get_opponent_state)
                        return
            else:
                print(f"Failed to get opponent state: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.get_opponent_state)
                return
        elif response.get("response_type") == "game_ready":
            print("Received 'game_ready' during opponent state check. Retrying...")
            self.root.after(REQUEST_INTERVAL, self.get_opponent_state)
            return
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.get_opponent_state)

    def check_game_result(self):
        print("*** Checking game result ***")

        # Send a request to check the game result
        response = self.client.send_request("request_type=game_result", "game_result")
        print(f"Response in check_game_result: {response}")

        if response.get("response_type") == "game_result":
            if response.get("status_code") == "200":
                result_message = response.get("message").strip()
                print(f"Game result: {result_message}")

                self.check_result = True

                if result_message == "Win":
                    print("You win! Congratulations!")
                    result_message = "You win! Congratulations!"
                    self.root.after(0, lambda: self.update_labels(opponent_health=0))
                elif result_message == "Lose":
                    print("You lose! Better luck next time!")
                    result_message = "You lose! Better luck next time!"
                elif result_message == "Draw":
                    print("Game ended in a draw.")
                    result_message = "Game ended in a draw."
                    self.root.after(0, lambda: self.update_labels(opponent_health=0))

                # Update the labels
                self.root.after(0, lambda: self.update_labels(info_message=result_message))

                # Ask player to play again or quit
                self.root.after(REQUEST_INTERVAL, lambda: self.ask_play_again(result_message))
            elif response.get("status_code") == "400":
                print("Game is still running... Fetching opponent state.")
                self.get_opponent_state()
            else:
                print(f"Failed to get game result: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(REQUEST_INTERVAL, self.check_game_result)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.check_game_result)


    def check_game_state(self):

        if self.cancel_call:
            return

        print("*** Checking game state ***")


        # Send a request to check the game state
        response = self.client.send_request("request_type=game_state", "game_state")
        print(f"Response in check_game_state: {response}")

        if response.get("response_type") == "game_state":
            if response.get("status_code") == "200":
                state_message = response.get("message").strip()
                print(f"Game state: {state_message}")

                if state_message == "Exit":
                    print("Game has ended due to opponent exit.")
                    self.update_labels(info_message="Opponent has left the game. You win!")
                    self.root.after(CHECK_INTERVAL, lambda: self.ask_play_again("Opponent has left the game. You win!"))
                elif state_message == "Reconnect":
                    print("Game is reconnecting. Waiting for opponent...")
                    
                    # Disable actions while reconnecting
                    self.check_activity = False
                    self.after_reconnect = True
                    self.disable_actions(reconnect=True)

                    self.root.after(CHECK_INTERVAL, self.check_game_state)
                elif state_message == "Running":
                    print("Game is still running. Checking again shortly...")

                    # After reconnecting, enable actions
                    if self.after_reconnect:
                        print("Game is running after reconnect.")
                        self.check_activity = True
                        self.after_reconnect = False
                        self.enable_actions()

                    self.root.after(CHECK_INTERVAL, self.check_game_state)

                elif state_message == "Over":

                    if self.check_result:
                        print("Game is over. Checking results...")
                        self.check_game_result()

            else:
                print(f"Failed to get game state: {response.get('message')}")
                self.update_labels(error_message=response.get("message"))
                self.root.after(CHECK_INTERVAL, self.check_game_state)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(CHECK_INTERVAL, self.check_game_state)


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
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(REQUEST_INTERVAL, self.reset_game)

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
            self.root.after_cancel(self.get_player_state)
            self.root.after_cancel(self.get_opponent_state)
            self.root.after_cancel(self.check_game_result)
            self.root.after_cancel(self.check_round_state)
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

        # Close the client connection
        try:
            self.client.close()
        except Exception as e:
            print(f"Error closing client: {e}")
        else:
            print("Client closed successfully.")

        self.root.destroy()
