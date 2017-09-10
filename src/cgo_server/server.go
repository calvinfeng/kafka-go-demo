package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"path/filepath"
)

type Server struct {
	Upgrader        websocket.Upgrader
	StreamConsumers map[*websocket.Conn]*kafka.Consumer
	StreamProducers map[*websocket.Conn]*kafka.Producer
	Clients         map[*websocket.Conn]bool
	broadcast       chan []byte
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

func (s *Server) handleStreamConnection(writer httpResponseWriter, req *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	wsConn, err := s.Upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer wsConn.Close()

  s.Clients[wsConn] = true
}
