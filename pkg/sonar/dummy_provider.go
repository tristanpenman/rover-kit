package sonar

import (
	"context"
	"time"
)

type DummyProvider struct{}

func (p *DummyProvider) Open(context.Context) chan Reading {
	c := make(chan Reading)

	go func() {
		for {
			c <- Reading{
				DistanceCM: 0,
				DurationUS: 0,
				Timestamp:  time.Now(),
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	return c
}

func (p *DummyProvider) Close(context.Context) error {
	return nil
}
