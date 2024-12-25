import time
import socket
import threading
from const import BUFFER_SIZE, PING_INTERVAL, REQUEST_INTERVAL

class Client:


    def __init__(self):
        self.socket = None
        self.server_ip = None
        self.server_port = None
        self.lock = threading.Lock()


    def connect(self):
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.connect((self.server_ip, self.server_port))
            print(f"Connected to server at {self.server_ip}:{self.server_port}")

            threading.Thread(target=self.ping, daemon=True).start()

            return True
        except Exception as e:
            print(f"Failed to connect to server: {e}")
            self.socket = None
            return False
    

    def set_server_info(self, ip, port):
        self.server_ip = ip
        self.server_port = port


    def send_request(self, request, request_type):
        with self.lock:
            try:
                if not self.socket:
                    raise ConnectionError("Not connected to the server.")

                print(f"Sending request: {request}")
                self.socket.sendall(request.encode('utf-8'))

                # Wait for the correct response type
                while True:
                    response_data = self.socket.recv(BUFFER_SIZE).decode('utf-8')
                    # print(f"Received raw response: {response_data}")

                    if not response_data:
                        print("No data received. Retrying...")
                        time.sleep(REQUEST_INTERVAL)  # Wait briefly before retrying
                        continue
                
                    response = self.parse_response(response_data)
                    print(f"Parsed response: {response}")

                    if response.get("response_type") == request_type:
                        return response

                    print(f"Ignored unrelated response: {response}")
                    time.sleep(REQUEST_INTERVAL)
            except Exception as e:
                print(f"Error sending request: {e}")
                # self.connect()
                return {"status": "error", "message": str(e)}


    def parse_response(self, response_data):
        response = {}
        pairs = response_data.split("&")
        for pair in pairs:
            if "=" not in pair:
                print(f"Malformed pair ignored: {pair}")
                continue
            key, value = pair.split("=", 1)
            response[key] = value
        return response
    

    def ping(self):
        while True:
            try:
                response = self.send_request("request_type=ping", "ping")
                print(f"Ping response: {response}")
            except Exception as e:
                print(f"Ping failed: {e}")
                break
            time.sleep(PING_INTERVAL)


    def close(self):
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
