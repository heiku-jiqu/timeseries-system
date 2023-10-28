package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type JSONpayload struct {
	Type      string   `json:"type"`
	ProductID []string `json:"product_id"`
	Channels  []string `json:"channels"`
}

type Ticker struct {
	Type        string    `json:"type"`
	Sequence    int       `json:"sequence"`
	ProductID   string    `json:"product_id"`
	Price       float32   `json:"price,string"`
	Open24H     float32   `json:"open_24h,string"`
	Volume24H   float32   `json:"volume_24h,string"`
	Low24H      float32   `json:"low_24,string"`
	High24H     float32   `json:"high_24,string"`
	Volume30D   float32   `json:"volume_30d,string"`
	BestBid     float32   `json:"best_bid,string"`
	BestBidSize float32   `json:"best_bid_size,string"`
	BestAsk     float32   `json:"best_ask,string"`
	BestAskSize float32   `json:"best_ask_size,string"`
	Side        string    `json:"side"`
	Time        time.Time `json:"time"`
	TradeID     int       `json:"trade_id"`
	LastSize    float32   `json:"last_size,string"`
}

func ParseTickerJSON(msg []byte) (Ticker, error) {
	ticker := Ticker{}
	err := json.Unmarshal(msg, &ticker)
	if err != nil {
		return Ticker{}, err
	}
	return ticker, nil
}

var (
	websocketURL        string   = "wss://ws-feed.exchange.coinbase.com"
	subscriptionChannel string   = "ticker"
	productIDs          []string = []string{"ETH-USD"}
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
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
				log.Printf("recv: parsed ticker: %v", ticker)
			} else {
				log.Printf("recv: %s", message)
			}
		}
	}()

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
}`, subscriptionChannel, `"`+strings.Join(productIDs, `","`)+`"`))
	fmt.Printf("%s", string(jsonPayload))

	err = c.WriteMessage(websocket.TextMessage, jsonPayload)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
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
