package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"syscall"
	"os/signal"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <group> <topics..>\n", os.Args[0])
		os.Exit(1)
	}

	// Expect at least 3 arguments from the command line input
	broker := os.Args[1]
	group := os.Args[2]
	topics := os.Args[3:]

	sigchan := make(chan os.Signal, 1) // Buffered channel
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    broker,
		"group.id":             group,
		"session.timeout.ms":   6000,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	run := true

	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("[Kafkapo][Message on %s]: %s\n",  e.TopicPartition, string(e.Value))
			case kafka.PartitionEOF:
				// fmt.Printf("%% Reached %v\n", e)
			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
				run = false
			default:
				// fmt.Printf("Ignored %v\n", e)

			}
		}
	}
}
