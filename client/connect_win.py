import tkinter as tk
from client import Client
from menu_win import MenuWindow
from const import SERVER_IP, SERVER_PORT, WIN_SIZE, TITLE_FONT, BUTTON_FONT, INPUT_FONT, ERROR_FONT

class ConnectWindow:

    def __init__(self, root=None):
        self.client = Client()

        if root is None:
            self.root = tk.Tk()
        else:
            self.root = root

        self.root.title("Connect")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)

        # Title
        tk.Label(self.root, text="Connect to the Server", font=TITLE_FONT).pack(pady=20)

        # Server input field
        tk.Label(self.root, text="Server IP:", font=INPUT_FONT).pack(pady=5)
        self.server_ip_entry = tk.Entry(self.root, font=INPUT_FONT)
        self.server_ip_entry.pack(pady=5)
        self.server_ip_entry.insert(0, SERVER_IP)

        # Server port input field
        tk.Label(self.root, text="Server Port:", font=INPUT_FONT).pack(pady=5)
        self.server_port_entry = tk.Entry(self.root, font=INPUT_FONT)
        self.server_port_entry.pack(pady=5)
        self.server_port_entry.insert(0, SERVER_PORT)

        # Error feedback
        self.error_label = tk.Label(self.root, text="", font=ERROR_FONT, fg="red")
        self.error_label.pack(pady=10)

        # Buttons
        tk.Button(self.root, text="Connect", font=BUTTON_FONT, command=self.connect_and_start).pack(pady=10)
        tk.Button(self.root, text="Quit", font=BUTTON_FONT, command=self.quit).pack(pady=10)

    def connect_and_start(self):
        try:
            server_ip = self.server_ip_entry.get().strip()
            server_port = self.server_port_entry.get().strip()

            if not server_ip or not server_port:
                print("Please enter a valid server IP and port.")
                self.error_label.config(text="Please enter a valid server IP and port.")
                return

            self.client.set_server_info(server_ip, int(server_port))

            if self.client.connect():
                self.load_menu_window()
            else:
                print("Failed to connect to the server.")
                self.error_label.config(text="Failed to connect to the server.")
        except ValueError:
            print("Invalid port number. Please enter a valid integer.")
            self.error_label.config(text="Invalid port number. Please enter a valid integer.")
        except Exception as e:
            print(f"An error occurred: {e}")
            self.error_label.config(text=f"An error occurred: {e}")

    def load_menu_window(self):
        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Load the menu interface
        MenuWindow(self.root, self.client)


    def quit(self):
        self.client.close()
        self.root.destroy()


    def run(self):
        self.root.mainloop()