package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
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

// Receive Tickers from kafkaChan until it closes or ctx is done
func (k *TickerKafka) ReceiveAndFlush(ctx context.Context, kafkaChan <-chan Ticker) {
	buf := make([]Ticker, 10)
	repeater := time.NewTicker(2 * time.Second) // flush interval

	for t := range kafkaChan {
		select {
		case <-ctx.Done():
			return

		default:
			// Keep appending new values to buffer
			// and flush the buffer to Kafka every 2 seconds
			buf = append(buf, t)
			select {
			case <-repeater.C:
				err := k.Write(ctx, buf...)
				if err != nil {
					log.Fatal().Err(err).Send()
				}
				buf = buf[:0]
			default:
			}
		}
	}
}
