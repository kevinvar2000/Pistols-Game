import time
import socket
import threading
from const import BUFFER_SIZE, PING_INTERVAL, REQUEST_INTERVAL, MAX_RETRIES

class Client:

    def __init__(self):
        self.socket = None
        self.server_ip = None
        self.server_port = None
        self.lock = threading.Lock()  # Lock to ensure thread safety

    def connect(self):
        try:
            # Create a new socket using the given address family and socket type
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            # Connect the socket to the server using the provided IP and port
            self.socket.connect((self.server_ip, self.server_port))
            print(f"Connected to server at {self.server_ip}:{self.server_port}")

            # Send a hello message to the server
            if not self.send_hello():
                return False

            # Start a new thread to send periodic ping requests to the server
            threading.Thread(target=self.ping, daemon=True).start()

            return True
        except Exception as e:
            print(f"Failed to connect to server: {e}")
            # If connection fails, set the socket to None
            self.socket = None
            return False

    def set_server_info(self, ip, port):
        # Set the server IP and port
        self.server_ip = ip
        self.server_port = port


    def send_hello(self):
        # Send a hello message to the server
        response = self.send_request("request_type=hello", "hello")
        print(f"Hello response: {response}")

        if response.get("response_type") == "hello":
            if response.get("status_code") == "200":
                print("Server accepted the connection.")
                print(f"Server message: {response.get('message')}")
                return True
            else:
                print("Server rejected the connection.")
                self.close()
                return False
        else:
            print("Unexpected response while connecting to the server.")
            self.close()
            return False


    def send_request(self, request, request_type):
        retries = 0

        with self.lock:
            try:
                if not self.socket:
                    raise ConnectionError("Not connected to the server.")

                print(f"Sending request: {request}")
                self.socket.sendall(request.encode('utf-8'))

                # Wait for the correct response type
                while retries < MAX_RETRIES:
                    response_data = self.socket.recv(BUFFER_SIZE).decode('utf-8')
                    # print(f"Received raw response: {response_data}")

                    if not response_data:
                        print("No data received. Retrying...")
                        retries += 1
                        time.sleep(REQUEST_INTERVAL)  # Wait briefly before retrying
                        continue

                    response = self.parse_response(response_data)
                    # print(f"Parsed response: {response}")

                    if response.get("response_type") == request_type:
                        return response

                    print(f"Ignored unrelated response: {response}")
                    retries += 1
                    time.sleep(REQUEST_INTERVAL)

                    print(f"Retrying request {retries}/{MAX_RETRIES}...")

                    # Optionally, track unrelated responses for analysis
                    with open("unrelated_responses.log", "a") as log_file:
                        print(f"Ignored unrelated response: {response}", file=log_file)
                        log_file.write(f"{response}\n")

                # If maximum retries are exceeded
                raise TimeoutError("Did not receive the expected response in time.")
            except Exception as e:
                (f"Error sending request: {e}")
                self.connect()
                return {"status": "error", "message": str(e)}

    def parse_response(self, response_data):
        # Parse the response data into a dictionary
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
        # Periodically send ping requests to the server
        while True:
            try:
                response = self.send_request("request_type=ping", "ping")
                print(f"Ping response: {response}")
            except Exception as e:
                print(f"Ping failed: {e}")
                self.connect()
                break
            time.sleep(PING_INTERVAL)

    def close_game(self):
        # Send a request to close the connection with the server
        response = self.send_request("request_type=close_game", "close_game")
        print(f"Close response: {response}")

        if response.get("response_type") == "close_game":
            if response.get("status") == "200":
                print("Server closed the connection.")
            else:
                print("Failed to close the connection.")
        else:
            print("Unexpected response while closing the connection.")

    def close(self):
        # Close the socket connection
        if self.socket:
            self.socket.close()
            self.socket = None
            print("Connection closed.")
        else:
            print("No connection to close.")
