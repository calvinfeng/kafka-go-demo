package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

/*
	AsyncProducer, SyncProducer, and Consumer are interfaces from sarama package. When we call NewAsyncProducer,
	NewSyncProducer, and NewConsumer, the return values are actually a pointer to an internal struct.
*/

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

func (s *Server) GetRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/chat/streams", s.handleStreamConnection)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../../static")))
	// return s.withAccessLog(s.collectQueryStringData())
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

func (s *Server) Close() error {
	if err := s.DataCollector.Close(); err != nil {
		log.Println("Failed to shut down data collector cleanly", err)
	}

	if err := s.AccessLogProducer.Close(); err != nil {
		log.Println("Failed to shut down access log producer cleanly", err)
	}

	return nil
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

	whenRoutineExit := make(chan bool)
	go s.ConsumePartitionAndPublish(wsConn, "chat", whenRoutineExit) // Since we only have one partition, we will use 0 as default

	for {
		var msg Message
		_, messageBytes, err := wsConn.ReadMessage()
		json.Unmarshal(messageBytes, &msg) // Unloading message bytes into the struct

		if err != nil {
			log.Printf("Websocket error: %v, now closing websocket connection...", err)
			delete(s.Clients, wsConn)

			<- whenRoutineExit // Wait for partition consumption to exit before we proceed to close Producer/Consumer

			log.Print("Now closing StreamProducer...")
			err := s.StreamProducers[wsConn].Close()
			if err != nil {
				log.Printf("Failed to close Producer cleanly: %v", err)
			}

			log.Print("Now closing StreamConsumer...")
			err = s.StreamConsumers[wsConn].Close()
			if err != nil {
				log.Printf("Failed to close Consumer cleanly: %v", err)
			}
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

func (s *Server) ConsumePartitionAndPublish(ws *websocket.Conn, topic string, exit chan bool) {
	// pc stands for PartitionConsumer
	pc, err := s.StreamConsumers[ws].ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Printf("Failed to consume partition: %v", err)
	}

	defer pc.Close()

	for s.Clients[ws] {
		select {
		case err := <-pc.Errors():
			log.Printf("[Kafkapo Consumer][error]: %v", err)
		case consumerMsg := <-pc.Messages():
			log.Printf("[Kafkapo Consumer][%v][%v][%v]\n", consumerMsg.Topic, consumerMsg.Partition, consumerMsg.Offset)
			for client := range s.Clients {
				var msg Message
				json.Unmarshal(consumerMsg.Value, &msg) // Unloading message bytes into the struct
				if err := client.WriteJSON(msg); err != nil {
					log.Printf("error: %v", err)
				}
			}
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Println("Now closing PartitionConsumer...")
	exit <- true
}

func (s *Server) withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()

		next.ServeHTTP(w, r)

		entry := &accessLogEntry{
			Method:       r.Method,
			Host:         r.Host,
			Path:         r.RequestURI,
			IP:           r.RemoteAddr,
			ResponseTime: float64(time.Since(started)) / float64(time.Second),
		}

		// We will use the client's IP address as key. This will cause
		// all the access log entries of the same IP address to end up
		// on the same partition.
		s.AccessLogProducer.Input() <- &sarama.ProducerMessage{
			Topic: "access_log",
			Key:   sarama.StringEncoder(r.RemoteAddr),
			Value: entry,
		}
	})
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
		DataCollector:     newStreamProducer(brokerList),
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
