package main

import (
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	ch := make(chan Ticker)
	defer close(ch)

	calc := Calculated{
		Avg: make(Average),
	}
	go calc.Process(ch)

	ch <- Ticker{
		Type:        "ticker",
		Sequence:    68422014160,
		ProductID:   "BTC-USD",
		Price:       37514.82,
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
	t.Run("Calculates average", func(t *testing.T) {
		expect := float32(37514.82)
		got := calc.Avg["BTC-USD"]
		if got != expect {
			t.Errorf("expected: %v, got %v", expect, got)
		}
	})
}
