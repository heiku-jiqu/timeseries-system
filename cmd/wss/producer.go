package main

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type TickerKafka struct {
	w *kafka.Writer
}

func NewDefaultKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "coinbase-ticker",
		Balancer: kafka.CRC32Balancer{},
	}
}

func (k *TickerKafka) Write(ctx context.Context, tickers ...Ticker) error {
	payloads := make([]kafka.Message, len(tickers))
	for i, tick := range tickers {
		payload, err := json.Marshal(tick)
		if err != nil {
			return err
		}
		payloads[i] = kafka.Message{
			Key:   []byte(tick.ProductID),
			Value: payload,
		}
	}
	err := k.w.WriteMessages(ctx, payloads...)
	if err != nil {
		return err
	}
	return nil
}
