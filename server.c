#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>

#define MAX_CLIENTS 10

void handleClient(int clientSocket) {
    char buffer[1024] = {0};
    int valread;
    
    while ((valread = read(clientSocket, buffer, sizeof(buffer))) > 0) {
        printf("Client message: %s\n", buffer);
        // Zde by následovala logika hry na základě zprávy od klienta
        // Včetně zasílání zpráv klientům a aktualizace stavu hry
        memset(buffer, 0, sizeof(buffer));
    }
}

int main() {
    int serverSocket, clientSocket;
    struct sockaddr_in serverAddr, clientAddr;
    socklen_t addrLen = sizeof(struct sockaddr);

    serverSocket = socket(AF_INET, SOCK_STREAM, 0);
    if (serverSocket == -1) {
        perror("Socket creation failed");
        exit(EXIT_FAILURE);
    }

    serverAddr.sin_family = AF_INET;
    serverAddr.sin_addr.s_addr = INADDR_ANY;
    serverAddr.sin_port = htons(12345); // Port

    if (bind(serverSocket, (struct sockaddr *)&serverAddr, sizeof(serverAddr)) < 0) {
        perror("Binding failed");
        exit(EXIT_FAILURE);
    }

    if (listen(serverSocket, MAX_CLIENTS) < 0) {
        perror("Listening failed");
        exit(EXIT_FAILURE);
    }

    while (1) {
        clientSocket = accept(serverSocket, (struct sockaddr *)&clientAddr, &addrLen);
        if (clientSocket < 0) {
            perror("Accepting connection failed");
            exit(EXIT_FAILURE);
        }

        int pid = fork(); // Vytvoření nového procesu pro obsluhu klienta
        if (pid < 0) {
            perror("Fork failed");
            exit(EXIT_FAILURE);
        }

        if (pid == 0) {
            close(serverSocket); // Zavření serverového socketu v potomkovi
            handleClient(clientSocket);
            close(clientSocket); // Zavření socketu klienta po obsloužení
            exit(EXIT_SUCCESS);
        } else {
            close(clientSocket); // Zavření socketu v rodiči
        }
    }

    close(serverSocket); // Toto by se nikdy nemělo stát v kódu
    return 0;
}
