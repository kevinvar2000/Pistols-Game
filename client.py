import socket
import threading

class Client:

    def __init__(self):
    
        self.server_address = ('127.0.0.1', 8080)  # IP adress and port of the server
        self.client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.client_socket.connect(self.server_address)

        # Thread for receiving messages from the server
        receive_thread = threading.Thread(target=self.receive_messages)
        receive_thread.start()


    def receive_messages(self):

        while True:
            try:
                message = self.client_socket.recv(1024).decode('utf-8')
                if message:
                    print(f"Server: {message}")
            except Exception as e:
                print(f"Error receiving message: {e}")
                break

    def send_message(self, message):

        try:
            self.client_socket.sendall(message.encode('utf-8'))
        except Exception as e:
            print(f"Error sending message: {e}")

if __name__ == "__main__":
    
    client = Client()
    
    while True:
        user_input = input("Select an action [SHOOT/RELOAD/COVER] or type 'EXIT' to quit: ")
        client.send_message(f"PLAYER_ACTION: {user_input}")
