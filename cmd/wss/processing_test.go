package main

import (
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	calc := NewCalculated()
	ch := make(chan Ticker)
	go calc.Process(ch)

	price := float32(1000)
	sampleTicker := Ticker{
		Type:        "ticker",
		Sequence:    68422014160,
		ProductID:   "BTC-USD",
		Price:       price,
		Open24H:     35586.77,
		Volume24H:   20560.37728398,
		Low24H:      35586.76,
		High24H:     37987,
		Volume30D:   395876.65432201,
		BestBid:     37512.01,
		BestBidSize: 0.02036921,
		BestAsk:     37514.82,
		BestAskSize: 0.17618076,
		Side:        "buy",
		Time:        time.Date(2023, 11, 16, 3, 51, 21, 0, time.UTC),
		TradeID:     577961606,
		LastSize:    0.00131924,
	}
	ch <- sampleTicker
	close(ch)
	t.Run("Calculates average, first entry", func(t *testing.T) {
		expect := price
		got := calc.GetAvg("BTC-USD")
		if got != expect {
			t.Errorf("expected: %v, got %v", expect, got)
		}
	})

	sampleTickerNext := sampleTicker
	sampleTickerNext.Price = 0.0
	ch = make(chan Ticker)
	go calc.Process(ch)
	ch <- sampleTickerNext
	close(ch)
	t.Run("Calculates average subsequent entries", func(t *testing.T) {
		expect := price / 2
		got := calc.GetAvg("BTC-USD")
		if got != expect {
			t.Errorf("expected: %v, got %v", expect, got)
		}
	})
}
