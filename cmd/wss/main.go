package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	qdb "github.com/questdb/go-questdb-client/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type JSONpayload struct {
	Type      string   `json:"type"`
	ProductID []string `json:"product_id"`
	Channels  []string `json:"channels"`
}

var (
	websocketURL        string   = "wss://ws-feed.exchange.coinbase.com"
	subscriptionChannel *string  = flag.String("channel", "ticker", "Specify `channel` to listen to. One of ticker or ticker_batch.")
	qdbAddr             *string  = flag.String("qdb", "127.0.0.1:9009", "Specify `url` of QuestDB's Influx Line Protocol")
	productIDs          []string = []string{"ETH-USD", "BTC-USD"}
)

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		productIDs = flag.Args()
	}
	// OS Interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Setup QuestDB connection
	checkQdbILPConn(*qdbAddr)
	ctx := context.Background()
	sender, err := qdb.NewLineSender(ctx, qdb.WithAddress(*qdbAddr))
	defer sender.Flush(ctx)
	defer sender.Close()

	// Setup WebSocket
	c, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Error().Err(err).Msg("error dialing wss")
	}
	defer c.Close()
	ws := &webSocketDatastream{c}

	// Setup waitgroup to wait for all goroutines before exiting main
	wg := sync.WaitGroup{}
	defer wg.Wait()

	done, tickerChan := ws.receiveDatastream()

	// Setup broadcasting from tickerChan to downstream channels
	qdbChan := make(chan Ticker, 1)
	kafkaChan := make(chan Ticker, 10)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(qdbChan)
		defer close(kafkaChan)
		for t := range tickerChan {
			// log.Printf("qdbChan: %v, kafkaChan: %v", len(qdbChan), len(kafkaChan))
			qdbChan <- t
			kafkaChan <- t
		}
	}()

	// Setup QuestDB Producer
	flushTicker := time.NewTicker(2 * time.Second)
	defer flushTicker.Stop()
	tickerModel := TickerModel{sender, flushTicker}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range qdbChan {
			err := tickerModel.Insert(t)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		}
	}()

	// Setup Kafka Producer
	kafkaWriter := NewDefaultKafkaWriter()
	defer kafkaWriter.Close()
	tickerKafka := TickerKafka{kafkaWriter}
	wg.Add(1)
	go func() {
		defer wg.Done()
		tickerKafka.ReceiveAndFlush(ctx, kafkaChan)
	}()

	// Setup Kafka Consumer
	consumer := NewDefaultKafkaConsumer()
	defer consumer.Close()
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Start(consumerCtx)
	}()

	// Start WebSocket data feed with data provider
	err = ws.initDatastream()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	// Block until cancelled or WebSocket connection ends
	ws.waitForInterrupt(done, interrupt)
}

func checkQdbILPConn(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to QuestDB InfluxLineProtocol")
	}
	defer conn.Close()
}
