package main

import (
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
	"encoding/json"
)

var (
	port      = flag.String("port", ":8000", "The address to bind to")
	brokers   = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose   = flag.Bool("verbose", false, "Turn on Sarama logging")
	certFile  = flag.String("certificate", "", "The optional certificate file for client authentication")
	keyFile   = flag.String("key", "", "The optional key file for client authentication")
	caFile    = flag.String("ca", "", "The optional certificate authority file for TLS client authentication")
	verifySsl = flag.Bool("verify", false, "Optional verify ssl certificates chain")
)

type Server struct {
	AccessLogProducer sarama.AsyncProducer
	DataCollector     sarama.SyncProducer
	Upgrader          websocket.Upgrader
	StreamConsumers   map[*websocket.Conn]sarama.Consumer
	StreamProducers   map[*websocket.Conn]sarama.SyncProducer
	Clients           map[*websocket.Conn]bool
	BrokerList        []string
}

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Hash     string `json:"hash"`
}

func (s *Server) Run(port string) error {
	httpServer := &http.Server{
		Addr:    port,
		Handler: s.GetRouter(),
	}

	log.Printf("Listening for requests on %s...\n", port)
	return httpServer.ListenAndServe()
}

func (s *Server) Close() error {
	if err := s.DataCollector.Close(); err != nil {
		log.Println("Failed to shut down data collector cleanly", err)
	}

	if err := s.AccessLogProducer.Close(); err != nil {
		log.Println("Failed to shut down access log producer cleanly", err)
	}

	return nil
}

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/chat/streams", s.handleStreamConnection)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../../static")))
	// return s.withAccessLog(s.collectQueryStringData())
	return r
}

func (s *Server) handleStreamConnection(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	wsConn, err := s.Upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer wsConn.Close()

	s.Clients[wsConn] = true
	s.StreamProducers[wsConn] = newStreamProducer(s.BrokerList)
	s.StreamConsumers[wsConn] = newStreamConsumer(s.BrokerList)
	go s.PublishToClient(wsConn, "chat")

	for {
		var msg Message
		_, messageBytes, err := wsConn.ReadMessage()
		json.Unmarshal(messageBytes, &msg) // Unloading message bytes into the struct

		if err != nil {
			log.Printf("Websocket error: %v", err)
			delete(s.Clients, wsConn)
			break
		}

		log.Println("Message arrived from socket client:", msg.Message, "from", msg.Username)
		partition, offset, err := s.StreamProducers[wsConn].SendMessage(&sarama.ProducerMessage{
		Topic: "chat",
			Value: sarama.ByteEncoder(messageBytes),
		})

		log.Printf("Message at produced at [%v][%v]", partition, offset)
	}
}

func (s *Server) PublishToClient(ws *websocket.Conn, topic string) {
	// pc stands for PartitionConsumer
	pc, err := s.StreamConsumers[ws].ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Failed to consume partition: %v", err)
	}

	listening := true
	for listening{
		select {
		case err := <- pc.Errors():
			log.Printf("Shit went wrong in consumer %v", err)
			listening = false
		case consumerMsg := <- pc.Messages():
			fmt.Printf("[Kafkapo Consumer][%v][%v][%v]\n", consumerMsg.Topic, consumerMsg.Partition, consumerMsg.Offset)
			for client := range s.Clients {
				var msg Message
				json.Unmarshal(consumerMsg.Value, &msg) // Unloading message bytes into the struct
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(s.Clients, client)
				}
			}
		}
	}
}

func (s *Server) collectQueryStringData() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// We are not setting a message key, which means that all messages will be distributed randomly over the different
		// partitions
		partition, offset, err := s.DataCollector.SendMessage(&sarama.ProducerMessage{
			Topic: "important",
			Value: sarama.StringEncoder(r.URL.RawQuery),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to store your data: %s", err)
		} else {
			// The tuple (topic, partition, offset) can be used as a unique identifier for a message in a Kafka cluster
			fmt.Fprintf(w, "Your data is stored with unique identifier important/%d/%d", partition, offset)
		}
	})
}

func (s *Server) withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// started := time.Now()

		next.ServeHTTP(w, r)

		// Implement this later
		// entry := ...
	})
}

func main() {
	flag.Parse()

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	brokerList := strings.Split(*brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))

	server := &Server{
		AccessLogProducer: newAccessLogProducer(brokerList),
		DataCollector: newStreamProducer(brokerList),
		StreamConsumers:   make(map[*websocket.Conn]sarama.Consumer),
		StreamProducers:   make(map[*websocket.Conn]sarama.SyncProducer),
		Clients:           make(map[*websocket.Conn]bool),
		BrokerList:        brokerList,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	defer func() {
		if err := server.Close(); err != nil {
			log.Println("Failed to close server", err)
		}
	}()

	log.Fatal(server.Run(*port))
}
