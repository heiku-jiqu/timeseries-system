package main

import (
	"context"
	"encoding/json"
	"time"

	qdb "github.com/questdb/go-questdb-client/v2"
)

type TickerModel struct {
	db *qdb.LineSender
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

func (t TickerModel) Insert(ticker Ticker) error {
	ctx := context.Background()
	err := t.db.Table("ticker").
		Symbol("type", ticker.Type).
		Symbol("product_id", ticker.ProductID).
		Symbol("side", ticker.Side).
		Int64Column("sequence", int64(ticker.Sequence)).
		Float64Column("price", float64(ticker.Price)).
		Float64Column("open_24h", float64(ticker.Open24H)).
		Float64Column("volume_24h", float64(ticker.Volume24H)).
		Float64Column("low_24", float64(ticker.Low24H)).
		Float64Column("high_24", float64(ticker.High24H)).
		Float64Column("volume_30d", float64(ticker.Volume30D)).
		Float64Column("best_bid", float64(ticker.BestBid)).
		Float64Column("best_bid_size", float64(ticker.BestBidSize)).
		Float64Column("best_ask", float64(ticker.BestAsk)).
		Float64Column("best_ask_size", float64(ticker.BestAskSize)).
		Int64Column("side", int64(ticker.TradeID)).
		Float64Column("last_size", float64(ticker.LastSize)).
		At(ctx, ticker.Time)
	if err != nil {
		return err
	}

	err = t.db.Flush(ctx)
	return err
}

func ParseTickerJSON(msg []byte) (Ticker, error) {
	ticker := Ticker{}
	err := json.Unmarshal(msg, &ticker)
	if err != nil {
		return Ticker{}, err
	}
	return ticker, nil
}
