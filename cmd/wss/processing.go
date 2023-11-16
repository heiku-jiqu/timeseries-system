package main

import "fmt"

type Calculated struct {
	Avg Average
}

// Key is ProductID, Value is the calculated average price.
type Average map[string]float32

func (c *Calculated) Process(tickerChan <-chan Ticker) {
	count := 0
	existTracker := make(map[string]struct{})
	for t := range tickerChan {
		count++
		if _, exist := existTracker[t.ProductID]; exist {
			c.Avg[t.ProductID] = c.Avg[t.ProductID] + (t.Price-c.Avg[t.ProductID])/float32(count)
		} else {
			c.Avg[t.ProductID] = t.Price
			existTracker[t.ProductID] = struct{}{}
		}
		fmt.Printf("consumer:\tcount %d\tavg %v\n", count, c.Avg)
	}
}
