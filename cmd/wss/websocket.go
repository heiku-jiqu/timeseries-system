package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type webSocketDatastream struct {
	c *websocket.Conn
}

// Listen for and parses incoming messages from `c`
// and pushes parsed message into into ch.
//
// `done` and `ch` channels will be closed when
// finished receiving from `c`.
func (ws *webSocketDatastream) receiveDatastream() (done chan struct{}, ch chan Ticker) {
	done = make(chan struct{})
	ch = make(chan Ticker, 200)
	go func() {
		// close channels after finish reading
		defer close(done)
		defer close(ch)
		for {
			_, message, err := ws.c.ReadMessage()
			if err != nil {
				log.Error().Err(err).Msg("read error")
				return
			}

			log.Debug().Msgf("recv: %s", message)
			if strings.Contains(string(message), `"type":"ticker"`) {
				ticker, err := ParseTickerJSON(message)
				if err != nil {
					log.Error().Err(err).Msg("parse error")
					return
				}
				ch <- ticker
			} else {
				log.Info().Str("payload", string(message)).Msg("received payload from wss")
			}
		}
	}()
	return done, ch
}

// Sends the initial websocket message to subscribe to tickers,
// causing the datastream to start
func (ws *webSocketDatastream) initDatastream() error {
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
	log.Printf("%s", string(jsonPayload))

	err := ws.c.WriteMessage(websocket.TextMessage, jsonPayload)
	return err
}

// Waits for signal from `done` or `interrupt` before returning.
// If `interrupted`, gracefully close `c`.
func (ws *webSocketDatastream) waitForInterrupt(done <-chan struct{}, interrupt <-chan os.Signal) {
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Print("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := ws.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error().Err(err).Msg("error closing wss")
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
