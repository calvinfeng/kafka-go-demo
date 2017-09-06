package main

import (
	"fmt"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

var topic string = "gossip"

const broker = "localhost:9092"
const consumerGroup = "bitch"

/*
A list of host/port pairs to use for establishing the initial connection to the Kafka cluster. The client will make use
of all servers irrespective of which servers are specified here for bootstrapping—this list only impacts the initial
hosts used to discover the full set of servers
*/

// Maps
var clients = make(map[*websocket.Conn]bool)
var producers = make(map[*websocket.Conn]*kafka.Producer)
var consumers = make(map[*websocket.Conn]*kafka.Consumer)

// Channels
var broadcast = make(chan Message)
var deliveryChan = make(chan kafka.Event)

// Socket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnection(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	// Upgrade initial GET request to a web socket connection
	wsConn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Close this when client is disconnected
	defer wsConn.Close()

	// Initialize a Kafka producer and consumer for this web socket stream
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})

	if err != nil {
		log.Fatal(err)
	}
	producers[wsConn] = p

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           consumerGroup,
		"session.timeout.ms": 6000,
		"enable.auto.commit": false,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
	consumers[wsConn] = c
	go wsListenForMessages(wsConn, []string{topic})

	// This will run as its own go-routine
	clients[wsConn] = true
	for {
		var msg Message
		_, messageBytes, err := wsConn.ReadMessage()
		json.Unmarshal(messageBytes, &msg) // Unloading message bytes into the struct

		if err != nil {
			log.Printf("Websocket error: %v", err)
			delete(clients, wsConn)
			break
		}

		log.Println("Message arrived from socket client:", msg.Message, "from", msg.Username)

		// Start producing messages
		err = p.Produce(
			&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Value: []byte(messageBytes),
			},
			deliveryChan,
		)

		// broadcast <- msg
	}
}

func wsListenForMessages(ws *websocket.Conn, topics []string) {
	c := consumers[ws]
	err := c.SubscribeTopics(topics, nil)
	listening := true

	if err != nil {
		log.Fatal(err)
		listening = false
	}

	for listening {
		event := c.Poll(100)
		if event == nil {
			continue
		}
		switch e := event.(type) {
		case *kafka.Message:
			fmt.Printf("[Kafkapo][Message on %s]: %s\n", e.TopicPartition, string(e.Value))
			for client := range clients {
				var msg Message
				json.Unmarshal(e.Value, &msg) // Unloading message bytes into the struct
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		case kafka.PartitionEOF:
			fmt.Println("Reached EOF, pending for more messages")
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			listening = false
		default:
			fmt.Printf("Ignored %v\n", e)
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
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
	http.HandleFunc("/chat/streams", handleConnection)

	go handleBroadcast()

	log.Println("Starting server on port 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe encountered an error: ", err)
	}
}
