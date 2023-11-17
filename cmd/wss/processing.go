package main

import (
	"log"
	"sync"
)

// Calculated holds calculated statistics.
// Calculated.Process(channel) to start processing events from the channel.
// Calculated.GetAvg() to get the average statistic.
type Calculated struct {
	avg          Average
	count        Count
	mu           sync.Mutex
	existTracker map[string]struct{}
}

// Key is ProductID, Value is the calculated average price.
type (
	Average map[string]float32
	Count   map[string]int
)

// NewCalculated creates a new Calcualted struct
func NewCalculated() *Calculated {
	return &Calculated{
		avg:          make(Average),
		count:        make(Count),
		existTracker: make(map[string]struct{}),
	}
}

// Get the productID's current price average.
// Thread safe.
func (c *Calculated) GetAvg(productID string) float32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.avg[productID]
	return out
}

// Process starts consuming Tickers from tickerChan
// and updates Calculated.Avg continuously until tickerChan is closed.
func (c *Calculated) Process(tickerChan <-chan Ticker) {
	for t := range tickerChan {
		c.process(t)
	}
}

// Function that updates statistics based on one Ticker
func (c *Calculated) process(t Ticker) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, exist := c.existTracker[t.ProductID]
	c.count[t.ProductID]++

	if exist {
		c.avg[t.ProductID] = c.avg[t.ProductID] + (t.Price-c.avg[t.ProductID])/float32(c.count[t.ProductID])
	} else {
		c.avg[t.ProductID] = t.Price
		c.existTracker[t.ProductID] = struct{}{}
	}
	log.Printf("consumer:\tcount %v\tavg %v\texist %v\n", c.count, c.avg, c.existTracker)
}
