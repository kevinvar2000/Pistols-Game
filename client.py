import socket
import threading

class Client:

    def __init__(self):
    
        self.server_address = ('127.0.0.1', 8080)  # IP adresa a port serveru
        self.client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.client_socket.connect(self.server_address)

        # Vlákno pro příjem zpráv ze serveru
        receive_thread = threading.Thread(target=self.receive_messages)
        receive_thread.start()

        # Uživatelské rozhraní nebo jiná logika klienta může být implementována zde

    def receive_messages(self):

        while True:
            try:
                message = self.client_socket.recv(1024).decode('utf-8')
                if message:
                    # Zpracování přijatých zpráv od serveru
                    print(f"Server: {message}")
            except Exception as e:
                print(f"Chyba při příjmu zprávy: {e}")
                break

    def send_message(self, message):

        try:
            self.client_socket.sendall(message.encode('utf-8'))
        except Exception as e:
            print(f"Chyba při odesílání zprávy: {e}")

if __name__ == "__main__":
    
    client = Client()
    
    while True:
        user_input = input("Zadejte akci (RELOAD/SHOOT/COVER): ")
        client.send_message(f"PLAYER_ACTION: {user_input}")
