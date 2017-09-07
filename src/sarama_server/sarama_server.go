package main

import (
	"flag"
	"github.com/Shopify/sarama"
	"log"
	"net/http"
	"os"
	"strings"
	"fmt"
)

var (
	port      = flag.String("port", ":8080", "The address to bind to")
	brokers   = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose   = flag.Bool("verbose", false, "Turn on Sarama logging")
	certFile  = flag.String("certificate", "", "The optional certificate file for client authentication")
	keyFile   = flag.String("key", "", "The optional key file for client authentication")
	caFile    = flag.String("ca", "", "The optional certificate authority file for TLS client authentication")
	verifySsl = flag.Bool("verify", false, "Optional verify ssl certificates chain")
)

type Server struct {
	DataCollector     sarama.SyncProducer
	AccessLogProducer sarama.AsyncProducer
}

func (s *Server) Run(port string) error {
	httpServer := &http.Server{
		Addr:    port,
		Handler: s.Handler(),
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

func (s *Server) Handler() http.Handler {
	return s.withAccessLog(s.collectQueryStringData())
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
		DataCollector:     newDataCollector(brokerList),
		AccessLogProducer: newAccessLogProducer(brokerList),
	}

	defer func() {
		if err := server.Close(); err != nil {
			log.Println("Failed to close server", err)
		}
	}()

	log.Fatal(server.Run(*port))
}