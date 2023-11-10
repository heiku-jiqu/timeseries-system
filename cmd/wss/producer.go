package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Creates new Producer. Caller needs to Flush() and Close() the Producer when done using.
func NewKafkaProducer() *kafka.Producer {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "coinbase-ticker-app",
	}
	p, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

// Logs Producer's message delivery
// Run as goroutine to not block
func deliveryReporter(p *kafka.Producer) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} else {
				fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
			}
		}
	}
}

func produceMessage(p *kafka.Producer, t Ticker) {
	topic := "coinbase-ticker"
	val, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(t.ProductID),
		Value:          val,
	}, nil)
}
