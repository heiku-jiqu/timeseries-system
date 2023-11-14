package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	qdb "github.com/questdb/go-questdb-client/v2"
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
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	checkQdbILPConn(*qdbAddr)
	ctx := context.TODO()
	sender, err := qdb.NewLineSender(ctx, qdb.WithAddress(*qdbAddr))
	defer sender.Flush(ctx)
	defer sender.Close()

	c, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	flushTicker := time.NewTicker(2 * time.Second)
	defer flushTicker.Stop()
	tickerModel := TickerModel{sender, flushTicker}
	kafkaWriter := NewDefaultKafkaWriter()
	defer kafkaWriter.Close()
	tickerKafka := TickerKafka{NewDefaultKafkaWriter()}

	tickerChan := make(chan Ticker, 200)
	done := make(chan struct{})
	wg := sync.WaitGroup{}
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveDatastream(c, done, tickerChan)
	}()

	qdbChan := make(chan Ticker, 1)
	kafkaChan := make(chan Ticker, 10)

	wg.Add(1)
	go func() {
		defer close(qdbChan)
		defer close(kafkaChan)
		defer wg.Done()
		for t := range tickerChan {
			// log.Printf("qdbChan: %v, kafkaChan: %v", len(qdbChan), len(kafkaChan))
			qdbChan <- t
			kafkaChan <- t
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range qdbChan {
			err := tickerModel.Insert(t)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]Ticker, 10)
		repeater := time.NewTicker(2 * time.Second)

		for t := range kafkaChan {
			buf = append(buf, t)
			select {
			case <-repeater.C:
				err := tickerKafka.Write(ctx, buf...)
				if err != nil {
					log.Fatal(err)
				}
				buf = buf[:0]
			default:
			}
		}
	}()

	err = initDatastream(c)
	if err != nil {
		log.Fatal(err)
	}

	waitForInterrupt(c, done, interrupt)
}

func checkQdbILPConn(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("could not connect to QuestDB InfluxLineProtocol: %v", err)
	}
	defer conn.Close()
}

// Receives to incoming messaages from `c`
// Close done channel when finished receiving
// Close ch channel when finished receiving
func receiveDatastream(c *websocket.Conn, done chan<- struct{}, ch chan<- Ticker) {
	defer close(done)
	defer close(ch)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		log.Printf("recv: %s", message)
		if strings.Contains(string(message), `"type":"ticker"`) {
			ticker, err := ParseTickerJSON(message)
			if err != nil {
				log.Println("read:", err)
				return
			}
			ch <- ticker
		} else {
			log.Printf("recv: %s", message)
		}
	}
}

// Sends the initial websocket message to subscribe to tickers
func initDatastream(c *websocket.Conn) error {
	jsonPayload := []byte(fmt.Sprintf(`{
    "type": "subscribe",
    "channels": [
        {
            "name": %q,
            "product_ids": [
                %s
            ]
        }
    ]
}`, *subscriptionChannel, `"`+strings.Join(productIDs, `","`)+`"`))
	fmt.Printf("%s", string(jsonPayload))

	err := c.WriteMessage(websocket.TextMessage, jsonPayload)
	return err
}

// Waits for signal from `done` or `interrupt` before returning.
// If `interrupted`, gracefully close `c`.
func waitForInterrupt(c *websocket.Conn, done <-chan struct{}, interrupt <-chan os.Signal) {
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
