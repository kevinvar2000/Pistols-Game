import time
import tkinter as tk
from const import WIN_SIZE, TITLE_FONT, BUTTON_FONT, INPUT_FONT, LABEL_FONT

class GameWindow:
    def __init__(self, root, client, client_name):
        self.client = client
        self.game_started = False

        self.root = root
        self.root.title("Game")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)

        self.get_player_state()

        # Client name display
        self.name_label = tk.Label(self.root, text=f"Name: {client_name}", font=LABEL_FONT)
        self.name_label.pack(pady=10)

        # Health and Ammo display
        self.health_label = tk.Label(self.root, text=f"Health: {self.health}", font=LABEL_FONT)
        self.health_label.pack(pady=10)

        self.ammo_label = tk.Label(self.root, text=f"Ammo: {self.ammo}", font=LABEL_FONT)
        self.ammo_label.pack(pady=10)

        # Buttons for actions
        tk.Button(self.root, text="Shoot", font=BUTTON_FONT, command=self.shoot).pack(pady=5)
        tk.Button(self.root, text="Reload", font=BUTTON_FONT, command=self.reload).pack(pady=5)
        tk.Button(self.root, text="Cover", font=BUTTON_FONT, command=self.cover).pack(pady=5)

        self.root.protocol("WM_DELETE_WINDOW", self.close_window)
        self.root.mainloop()

    def send_action(self, action):
        response = self.client.send_request(action)
        if response:
            self.update_game_state(response)

    def shoot(self):
        if self.ammo > 0:
            self.ammo -= 1
            self.send_action("request_type=action&action_type=shoot")
        else:
            print("No ammo left!")

    def reload(self):
        self.ammo = 5
        self.send_action("request_type=action&action_type=reload")

    def cover(self):
        self.send_action("request_type=action&action_type=cover")

    def update_game_state(self, state):
        try:
            # response_type=player_health&status_code=200&message=Health=3, Ammo=1
            if state.startswith("response_type=player_health"):
               if "Health" in state and "Ammo" in state:
                    self.health = int(state.split(",")[0].split("=")[1])
                    self.ammo = int(state.split(",")[1].split("=")[1])
                    self.health_label.config(text=f"Health: {self.health}")
                    self.ammo_label.config(text=f"Ammo: {self.ammo}")
                    self.root.update()
        except:
            print("Error updating game state.")


    def check_game_ready(self):
        while not self.game_started:
            print("Sending game_ready request...")
            response = self.client.send_game_ready()

            if response.get("response_type") == "game_ready" and response.get("status_code") == "200":
                print("Game is ready! Starting the game...")
                self.game_started = True
                self.load_game()
                break
            else:
                print("Game not ready yet. Retrying...")
                time.sleep(2)


    def get_player_state(self):
        self.client.send_request("request_type=player_state")
        response = self.client.receive_message()
        if response:
            self.update_game_state(response)

    def close_window(self):
        self.client.close()
        self.root.destroy()
