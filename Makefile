# Define OS-specific commands
OS := $(shell uname -s)

ifeq ($(OS),Linux)
    PYTHON_CMD = python3
else ifeq ($(OS),Darwin)
    PYTHON_CMD = python3
else
    PYTHON_CMD = python
endif

GO_CMD = go run .
GO_BUILD_CMD = go build -o bin/server .

# Initialize the Go module (only if needed)
init-go-mod:
	@if [ ! -f server/go.mod ]; then \
		cd server && go mod init server; \
	else \
		echo "Go module already initialized."; \
	fi

# Server target
server: init-go-mod
	@echo "Building the server binary..."
	@cd server && $(GO_BUILD_CMD)
	@echo "Binary built at bin/server."
	@echo "Running the server..."
	@if [ "$(filter-out server,$(MAKECMDGOALS))" != "" ]; then \
		echo "Starting the server with arguments: $(filter-out server,$(MAKECMDGOALS))"; \
		cd server && ./bin/server $(filter-out server,$(MAKECMDGOALS)); \
	else \
		cd server && ./bin/server; \
	fi

# Client target
client:
	@echo "Starting the client..."
	$(PYTHON_CMD) ./client/main.py

# Phony for the server and client targets
.PHONY: server client
