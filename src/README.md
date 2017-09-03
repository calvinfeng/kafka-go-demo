# Real-time Messaging Application
## Integrating with Kafka

First of all, add this project into your `GOPATH`
```
export GOPATH=/path/to/Kafkapo
```

Include GOPATH in your path
```
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
```

And then install Gorilla websocket & Kafka Golang client
```
cd /path/to/Kafkapo
go get github.com/gorilla/websocket
go get github.com/confluentinc/confluent-kafka-go/kafka
```
