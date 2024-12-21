import tkinter as tk
from game_win import GameWindow
from const import WIN_SIZE, TITLE_FONT, BUTTON_FONT, INPUT_FONT, ERROR_FONT, LABEL_FONT

class MenuWindow:


    def __init__(self, root, client):
        self.client = client

        self.root = root
        self.root.title("Menu")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)

        # Title
        tk.Label(self.root, text="Welcome to the Game!", font=TITLE_FONT).pack(pady=20)

        # Name input field
        tk.Label(self.root, text="Enter your name:", font=INPUT_FONT).pack(pady=5)
        self.name_entry = tk.Entry(self.root, font=INPUT_FONT)
        self.name_entry.pack(pady=5)

        # Error feedback
        self.error_label = tk.Label(self.root, text="", font=ERROR_FONT, fg="red")
        self.error_label.pack(pady=10)

        # Buttons
        tk.Button(self.root, text="Play", font=BUTTON_FONT, command=self.play).pack(pady=10)
        tk.Button(self.root, text="Back", font=BUTTON_FONT, command=self.back).pack(pady=10)


    def play(self):

        name = self.name_entry.get().strip()
        if not name or len(name) == 0:
            print("Please enter a valid name.")
            self.error_label.config(text="Please enter a valid name.")
            return
        
        response = self.client.send_request("request_type=name&name=" + name)

        print(f"Response in menu: {response}")

        if response.get("response_type") == "name":
            if response.get("status_code") == "200":
                self.load_game_window(name)
            else:
                print(f"Failed to set name: {response.get('message')}")
                self.error_label.config(text=response.get("message"))


    def load_game_window(self, name):
        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Load the game window
        GameWindow(self.root, self.client, name)


    def back(self):
        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        from connect_win import ConnectWindow

        # Load the connect
        ConnectWindow(self.root)


    def run(self):
        self.root.mainloop()