package main

import (
	"github.com/Shopify/sarama"
	"log"
)

func newStreamConsumer(brokerList []string) sarama.Consumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}

	return consumer
}
