package main

import "fmt"

type Calculated struct {
	Avg          Average
	Count        Count
	existTracker map[string]struct{}
}

// Key is ProductID, Value is the calculated average price.
type (
	Average map[string]float32
	Count   map[string]int
)

func NewCalculated() *Calculated {
	return &Calculated{
		Avg:          make(Average),
		Count:        make(Count),
		existTracker: make(map[string]struct{}),
	}
}

// Updates Calculated.Avg continuously until tickerChan is closed
func (c *Calculated) Process(tickerChan <-chan Ticker) {
	for t := range tickerChan {
		_, exist := c.existTracker[t.ProductID]
		fmt.Printf("%v\n\n", exist)
		c.Count[t.ProductID]++
		if exist {
			fmt.Printf("updating avg\n")
			c.Avg[t.ProductID] = c.Avg[t.ProductID] + (t.Price-c.Avg[t.ProductID])/float32(c.Count[t.ProductID])
		} else {
			fmt.Printf("init avg\n")
			c.Avg[t.ProductID] = t.Price
			c.existTracker[t.ProductID] = struct{}{}
		}
		fmt.Printf("consumer:\tcount %v\tavg %v\texist %v\n", c.Count, c.Avg, c.existTracker)
	}
}
