package main

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Hash     string `json:"hash"`
}

type Server struct {
	Upgrader        websocket.Upgrader
	StreamConsumers map[*websocket.Conn]*kafka.Consumer
	StreamProducers map[*websocket.Conn]*kafka.Producer
	Clients         map[*websocket.Conn]bool
	Broadcast       chan []byte
	Broker          string
}

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/chat/streams", s.handleStreamConnection)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(exPath + "../../static")))
	return r
}

func (s *Server) Run(port string) error {
	httpServer := &http.Server{
		Addr:    port,
		Handler: s.GetRouter(),
	}

	log.Printf("Listening for requests on %s...\n", port)
	return httpServer.ListenAndServe()
}

func (s *Server) handleStreamConnection(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	wsConn, err := s.Upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer wsConn.Close()

	s.Clients[wsConn] = true

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": s.Broker})

	if err != nil {
		log.Fatal(err)
	}
	s.StreamProducers[wsConn] = p

	consumerGroup := "Room-1"
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  s.Broker,
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
	s.StreamConsumers[wsConn] = c

	for {
		var msg Message
		_, messageBytes, err := wsConn.ReadMessage()

		if err != nil {
			log.Printf("Websocket error: %v", err)
			delete(s.Clients, wsConn)

			log.Println("Closing StreamProducer")
			s.StreamProducers[wsConn].Close()

			log.Println("Closing StreamConsumer")
			s.StreamConsumers[wsConn].Close()

			break
		}

		json.Unmarshal(messageBytes, &msg) // Unloading message bytes into the struct
		log.Printf("Message arrived from socket client: %v from %v", msg.Message, msg.Username)

		// Produce message asynchronously, but we will make it blocking until message production is confirmed for delivery
		topic := "chat"
		doneChan := make(chan bool)
		go s.confirmProducerDelivery(p, doneChan)
		p.ProduceChannel() <- &kafka.Message{ // This is a channel producer
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: messageBytes,
		}
		<-doneChan

		s.Broadcast <- messageBytes
	}
}

func (s *Server) confirmProducerDelivery(p *kafka.Producer, doneChan chan bool) {
	defer close(doneChan)
	// ei stands for Event Interface
	for ev := range p.Events() {
		switch event := ev.(type) {
		case *kafka.Message:
			msg := event
			if msg.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v\n", msg.TopicPartition.Error)
			} else {
				log.Printf("Delivered message to topic %s [%d] at offset %v\n",
					*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
			}
		default:
			log.Printf("Ignored event: %s\n", ev)
		}
	}
}

func (s *Server) HandleBroadcast() {
	for {
		msgBytes := <-s.Broadcast
		for client := range s.Clients {
			var msg Message
			json.Unmarshal(msgBytes, &msg) // Unloading message bytes into the struct
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(s.Clients, client)
			}
		}
	}
}

func (s *Server) clientConsumeMessages(wsConn *websocket.Conn, topics []string, signalExit chan bool) {
	c := s.StreamConsumers[wsConn]

	err := c.SubscribeTopics(topics, nil)
	if err != nil {
		log.Printf("Error in subscribing to topics: %v", err)
	}

	listening := true
	for listening {
		select {
		case sig := <-signalExit:
			log.Printf("Caught signal %v: terminating\n", sig)
			listening = false
		case ev := <-c.Events():
			switch event := ev.(type) {
			case kafka.AssignedPartitions: // Not exactly sure what are these for
				log.Printf("%v\n", event)
				c.Assign(event.Partitions)
			case kafka.RevokedPartitions: // Same here
				log.Printf("%v\n", event)
				c.Unassign()
			case *kafka.Message:
				log.Printf("Message on %s:\n%s\n", event.TopicPartition, string(event.Value))
				for client := range s.Clients {
					var msg Message
					json.Unmarshal(event.Value, &msg) // Unloading message bytes into the struct
					err := client.WriteJSON(msg)
					if err != nil {
						log.Printf("Failed to publish to client: %v", err)
						client.Close()
						delete(s.Clients, client)
					}
				}
			case kafka.PartitionEOF:
				log.Printf("Reached %v\n", event)
			case kafka.Error:
				log.Printf("Error: %v\n", event)
				listening = false
			}
		}
	}
}

func main() {
	server := &Server{
		StreamConsumers:   make(map[*websocket.Conn]*kafka.Consumer),
		StreamProducers:   make(map[*websocket.Conn]*kafka.Producer),
		Clients:           make(map[*websocket.Conn]bool),
		Broadcast:         make(chan []byte),
		Broker:            "localhost:9092",
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	log.Fatal(server.Run(":8000"))
}