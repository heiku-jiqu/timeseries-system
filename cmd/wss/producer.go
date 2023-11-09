package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Creates new Producer. Caller needs to Flush() and Close() the Producer when done using.
func NewKafkaProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func deliveryReporter(p *kafka.Producer) {
	go func() {
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
	}()
}

func produceMessage(p *kafka.Producer, t Ticker) {
	topic := "Cryptocurrency"
	val, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic},
		Key:            []byte(t.ProductID),
		Value:          val,
	}, nil)
}
