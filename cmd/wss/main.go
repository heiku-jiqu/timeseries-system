package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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
	productIDs          []string = []string{"ETH-USD"}
)

func main() {
	flag.Parse()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx := context.TODO()
	sender, err := qdb.NewLineSender(ctx)

	tickerModel := TickerModel{sender}

	c, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go receiveDatastream(c, done, func(t Ticker) {
		log.Printf("recv: parsed ticker: %v", t)
		err := tickerModel.Insert(t)
		if err != nil {
			log.Fatal(err)
		}
	})

	err = initDatastream(c)
	if err != nil {
		log.Fatal(err)
	}

	waitForInterrupt(c, done, interrupt)
}

// Receives to incoming messaages from `c`
// Sends done channel when finished receiving
func receiveDatastream(c *websocket.Conn, done chan<- struct{}, callback func(Ticker)) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		if strings.Contains(string(message), `"type":"ticker"`) {
			ticker, err := ParseTickerJSON(message)
			if err != nil {
				log.Println("read:", err)
				return
			}
			callback(ticker)
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
