package main

import (
	"encoding/json"
	"time"
)

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
