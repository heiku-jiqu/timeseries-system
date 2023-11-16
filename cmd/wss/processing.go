package main

import (
	"fmt"
	"sync"
)

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

// Updates Calculated.Avg continuously until tickerChan is closed
func (c *Calculated) Process(tickerChan <-chan Ticker) {
	for t := range tickerChan {
		_, exist := c.existTracker[t.ProductID]
		c.mu.Lock()
		defer c.mu.Unlock()
		c.count[t.ProductID]++

		if exist {
			c.avg[t.ProductID] = c.avg[t.ProductID] + (t.Price-c.avg[t.ProductID])/float32(c.count[t.ProductID])
		} else {
			c.avg[t.ProductID] = t.Price
			c.existTracker[t.ProductID] = struct{}{}
		}
		fmt.Printf("consumer:\tcount %v\tavg %v\texist %v\n", c.count, c.avg, c.existTracker)
	}
}
