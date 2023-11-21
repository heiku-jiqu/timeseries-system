package main

// Read Kafka -> Process (aggregate?) -> Output
//

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// Consumer that reads messages from coinbase-ticker channel
// published by producer.go
//
// Call Close() when finished to release resources
type KafkaConsumer struct {
	r  *kafka.Reader
	ch chan Ticker
}

func NewDefaultKafkaConsumer() *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     "coinbase-ticker-go-consumer",
		Topic:       "coinbase-ticker",
		StartOffset: kafka.LastOffset,
	})
	return &KafkaConsumer{
		r:  reader,
		ch: make(chan Ticker),
	}
}

func (k *KafkaConsumer) Close() error {
	log.Info().Msg("closing consumer")
	err := k.r.Close()
	return err
}

// Reads messages from the TickerKafka in producer.go
// Processes the data (aggregate)
// Writes to console the processed data
func (k *KafkaConsumer) Start(ctx context.Context) {
	// Read from kafka into channel k.ch
	go k.read(ctx)

	calc := NewCalculated()
	calc.Process(k.ch)
}

// Continuously reads and parse messages and sends into k.ch
func (k *KafkaConsumer) read(ctx context.Context) error {
	defer log.Print("closing channel")
	defer close(k.ch)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := k.r.FetchMessage(ctx)
			log.Debug().Str("kafka-message-val", string(msg.Value)).Msg("read Kafka message")
			if err != nil {
				return err
			}
			parsed, err := ParseTickerJSON(msg.Value)
			if err != nil {
				return err
			}

			k.ch <- parsed
		}
	}
}
