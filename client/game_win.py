import time
import tkinter as tk
from const import WIN_SIZE, TITLE_FONT, BUTTON_FONT, INPUT_FONT, LABEL_FONT, REQUEST_INTERVAL

class GameWindow:
    def __init__(self, root, client, client_name):
        self.client = client
        self.client_name = client_name

        self.root = root
        self.root.title("Game")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)

        # Start the waiting room until game is ready
        self.waiting_room()


    def send_action(self, action):
        response = self.client.send_request(action, "action")
        print(f"Response in send_action: {response}")

        if response.get("response_type") == "action":
            if response.get("status_code") == "200":
                self.get_player_state()
            else:
                print(f"Failed to send action: {response.get('message')}")
                self.error_label.config(text=response.get("message"))
                self.root.after(1000, lambda: self.send_action(action))
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(1000, lambda: self.send_action(action))


    def shoot(self):
        self.send_action("request_type=action&action_type=shoot")


    def reload(self):
        self.send_action("request_type=action&action_type=reload")


    def cover(self):
        self.send_action("request_type=action&action_type=cover")


    def waiting_room(self):
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
        print(f"Waiting for game to be ready...")
        
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
                self.root.after(1000, self.check_game_ready)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(1000, self.check_game_ready)


    def load_game(self):
        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()


        # Client name display
        self.name_label = tk.Label(self.root, text=f"Name: {self.client_name}", font=LABEL_FONT)
        self.name_label.pack(pady=10)

        # Health and Ammo display
        self.health_label = tk.Label(self.root, text="Health: Loading...", font=LABEL_FONT)
        self.health_label.pack(pady=10)

        self.ammo_label = tk.Label(self.root, text="Ammo: Loading...", font=LABEL_FONT)
        self.ammo_label.pack(pady=10)

        self.info_label = tk.Label(self.root, text="Choose an action:", font=LABEL_FONT)
        self.info_label.pack(pady=10)

        self.error_label = tk.Label(self.root, text="", font=LABEL_FONT, fg="red")
        self.error_label.pack(pady=10)


        # Get the initial player state
        self.get_player_state()


        # Buttons for actions
        tk.Button(self.root, text="Shoot", font=BUTTON_FONT, command=self.shoot).pack(pady=5)
        tk.Button(self.root, text="Reload", font=BUTTON_FONT, command=self.reload).pack(pady=5)
        tk.Button(self.root, text="Cover", font=BUTTON_FONT, command=self.cover).pack(pady=5)

        self.root.protocol("WM_DELETE_WINDOW", self.close_window)

        self.root.mainloop()


    def get_player_state(self):
        response = self.client.send_request("request_type=player_state", "player_state")
        print(f"Response in get_player_state: {response}")

        if response.get("response_type") == "player_state":
            if response.get("status_code") == "200":
                state_message = response.get("message").strip()
                print(f"Player state: {state_message}")

                if state_message == "Dead":
                    print("Player is dead. Ending game...")

                    # Update the labels
                    self.health_label.config(text="Health: 0")
                    self.ammo_label.config(text="Ammo: 0")
                    self.info_label.config(text="You are dead. Game over.")
                    self.root.update_idletasks()

                    print("Health label: ", self.health_label)
                    print("Ammo label: ", self.ammo_label)

                    return

                try:
                    state_pairs = dict(pair.split("=") for pair in state_message.split(", "))
                    self.health = int(state_pairs.get("Health", 0))
                    self.ammo = int(state_pairs.get("Ammo", 0))
                    print(f"Updated player state: Health={self.health}, Ammo={self.ammo}")

                    # Update the labels
                    self.health_label.config(text=f"Health: {self.health}")
                    self.ammo_label.config(text=f"Ammo: {self.ammo}")
                    self.root.update_idletasks()

                    print("Health label: ", self.health_label)
                    print("Ammo label: ", self.ammo_label)

                except (ValueError, KeyError) as e:
                    print(f"Error parsing player state: {e}")
                    self.error_label.config(text="Error parsing player state")

            else:
                print(f"Failed to get player state: {response.get('message')}")
                self.error_label.config(text=response.get("message"))
                self.root.after(1000, self.get_player_state)
        else:
            print(f"Ignoring unrelated response: {response}")
            self.root.after(1000, self.get_player_state)


    def close_window(self):
        self.client.close()
        self.root.destroy()
