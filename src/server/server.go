package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

const topic =

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

type Message struct {
    Email string `json:"email"`
    Username string `json:"username"`
    Message string `json:"message"`
}

func handleConnection(writer http.ResponseWriter, req *http.Request) {
    // Upgrade initial GET request to a websocket connection
    ws, err := upgrader.Upgrade(writer, req, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Close this when client is disconnected
    defer ws.Close()
    clients[ws] = true

    for {
        var msg Message
        // Read in new message as JSON and map it to a Message struct
        err := ws.ReadJSON(&msg)
        log.Println("Incoming message:", msg.Message)
        if err != nil {
            log.Printf("error: %v", err)
            delete(clients, ws)
            break
        }

        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <- broadcast
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func main() {
    fs := http.FileServer(http.Dir("../../static"))
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnection)

    go handleMessages()

    log.Println("Starting server on port 8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe encountered an error: ", err)
    }
}
