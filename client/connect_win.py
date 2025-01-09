import ipaddress
import tkinter as tk
from client import Client
from menu_win import MenuWindow
from const import SERVER_IP, SERVER_PORT, WIN_SIZE, TITLE_FONT, BUTTON_FONT, ERROR_FONT, ELEMENT_SIZE, WIN_BG, BTN_BG, BTN_FG, INPUT_BG, INPUT_FG, LABEL_FONT, LABEL_BG, LABEL_FG, ERROR_BG, ERROR_FG

class ConnectWindow:

    def __init__(self, root=None):
        # Initialize the client
        self.client = Client()

        # Initialize the root window
        if root is None:
            self.root = tk.Tk()
        else:
            self.root = root

        # Configure the root window
        self.root.title("Connect")
        self.root.geometry(WIN_SIZE)
        self.root.resizable(False, False)
        self.root.configure(bg=WIN_BG)

        # Title label
        tk.Label(self.root, text="Connect to the Server", font=TITLE_FONT, bg=LABEL_BG, fg=LABEL_FG).pack(pady=50)

        # Frame for input fields
        self.input_frame = tk.Frame(self.root, bg=WIN_BG)
        self.input_frame.pack(pady=10)

        # Server IP input field
        tk.Label(self.input_frame, text="Server IP:", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG).pack(pady=5)
        self.server_ip_entry = tk.Entry(self.input_frame, font=LABEL_FONT, width=ELEMENT_SIZE, bg=INPUT_BG, fg=INPUT_FG, justify="center")
        self.server_ip_entry.pack(pady=5)
        self.server_ip_entry.insert(0, SERVER_IP)

        # Server port input field
        tk.Label(self.input_frame, text="Server Port:", font=LABEL_FONT, bg=LABEL_BG, fg=LABEL_FG).pack(pady=5)
        self.server_port_entry = tk.Entry(self.input_frame, font=LABEL_FONT, width=ELEMENT_SIZE, bg=INPUT_BG, fg=INPUT_FG, justify="center")
        self.server_port_entry.pack(pady=5)
        self.server_port_entry.insert(0, SERVER_PORT)

        # Error feedback label
        self.error_label = tk.Label(self.input_frame, text="", font=ERROR_FONT, fg=ERROR_FG, bg=ERROR_BG, relief="solid")

        # Frame for buttons
        self.button_frame = tk.Frame(self.root, bg=WIN_BG)
        self.button_frame.pack(pady=10)

        # Connect button
        tk.Button(self.button_frame, text="Connect", font=BUTTON_FONT, command=self.connect_and_start, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=10)
        # Quit button
        tk.Button(self.button_frame, text="Quit", font=BUTTON_FONT, command=self.quit, width=ELEMENT_SIZE, bg=BTN_BG, fg=BTN_FG).pack(pady=10)

    def connect_and_start(self):
        try:
            # Get server IP and port from input fields
            server_ip = self.server_ip_entry.get().strip()
            server_port = self.server_port_entry.get().strip()

            # Validate input fields
            if not server_ip or not server_port:
                print("Please enter a valid server IP and port.")
                self.error_label.config(text="Please enter a valid server IP and port.")
                self.error_label.pack(pady=10)
                return

            # Validate IP address
            try:
                ipaddress.ip_address(server_ip)
            except ValueError:
                print("Invalid IP address. Please enter a valid IP.")
                self.error_label.config(text="Invalid IP address. Please enter a valid IP.")
                self.error_label.pack(pady=10)
                return

            # Validate port number
            try:
                server_port = int(server_port)
                if server_port < 1 or server_port > 65535:
                    raise ValueError("Port out of range")
            except ValueError:
                print("Invalid port number. Please enter a valid integer between 1 and 65535.")
                self.error_label.config(text="Invalid port number. Please enter a valid integer between 1 and 65535.")
                self.error_label.pack(pady=10)
                return

            # Set server info in the client
            self.client.set_server_info(server_ip, int(server_port))

            # Attempt to connect to the server
            if self.client.connect():
                self.load_menu_window()
            else:
                print("Failed to connect to the server.")
                self.error_label.config(text="Failed to connect to the server.")
                self.error_label.pack(pady=10)
        except Exception as e:
            print(f"An error occurred: {e}")
            self.error_label.config(text=f"An error occurred: {e}")
            self.error_label.pack(pady=10)

    def load_menu_window(self):
        # Clear the current window
        for widget in self.root.winfo_children():
            widget.destroy()

        # Load the menu interface
        MenuWindow(self.root, self.client)

    def quit(self):
        # Close the client and destroy the root window
        self.client.close()
        self.root.destroy()

    def run(self):
        # Run the Tkinter main loop
        self.root.mainloop()