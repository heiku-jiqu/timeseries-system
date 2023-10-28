package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type JSONpayload struct {
	Type      string   `json:"type"`
	ProductID []string `json:"product_id"`
	Channels  []string `json:"channels"`
}

const websocketURL string = "wss://ws-feed.exchange.coinbase.com"

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
			log.Printf("recv: %s", message)
		}
	}()

	// payload := JSONpayload{"subscribe", []string{"ETH-USD"}, []string{"ticker"}}
	// wsjson.Write(ctx, c, payload)
	jsonPayload := []byte(`{
    "type": "subscribe",
    "channels": [
        {
            "name": "ticker",
            "product_ids": [
                "ETH-USD"
            ]
        }
    ]
}`)
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
		// case t := <-ticker.C:
		// err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// if err != nil {
		// 	log.Println("write:", err)
		// 	return
		// }
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
