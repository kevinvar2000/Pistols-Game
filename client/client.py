import time
import socket
import threading
from const import BUFFER_SIZE

class Client:


    def __init__(self):
        self.socket = None
        self.server_ip = None
        self.server_port = None
        self.pinging = False


    def connect(self):
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.connect((self.server_ip, self.server_port))
            print(f"Connected to server at {self.server_ip}:{self.server_port}")

            self.pinging = True
            self.ping_thread = threading.Thread(target=self.ping, daemon=True).start()

            return True
        except Exception as e:
            print(f"Failed to connect to server: {e}")
            self.socket = None
            self.pinging = False
            return False
    

    def set_server_info(self, ip, port):
        self.server_ip = ip
        self.server_port = port


    def send_request(self, request):
        try:
            if not self.socket:
                raise ConnectionError("Not connected to the server.")

            print(f"Sending request: {request}")

            # Serialize the request as JSON
            self.socket.sendall(request.encode('utf-8'))

            # Receive the response
            response_data = self.socket.recv(BUFFER_SIZE).decode('utf-8')
            print(f"Response data: {response_data}")

            return  self.parse_response(response_data)
        except Exception as e:
            print(f"Error sending request: {e}")
            return {"status": "error", "message": str(e)}


    def parse_response(self, response_data):
        response = {}
        pairs = response_data.split("&")
        for pair in pairs:
            key, value = pair.split("=", 1)
            response[key] = value
        return response


    def ping(self):
        while self.pinging:
            try:
                response = self.send_request("request_type=ping")
                if response.get("response_type") != "ping" or response.get("status_code") != "200":
                    print("Failed to ping the server. Stopping ping.")
                    self.pinging = False
                    break
            except Exception as e:
                print(f"Ping failed: {e}")
                self.pinging = False
                break
            time.sleep(5)

    def send_game_ready(self):
        response = self.send_request("request_type=game_ready")
        return response

    def close(self):
        self.pinging = False
        if self.socket:
            try:
                self.socket.close()
                print("Connection closed.")
            except Exception as e:
                print(f"Error while closing the socket: {e}")
            finally:
                self.socket = None
        else:
            print("No connection to close.")
