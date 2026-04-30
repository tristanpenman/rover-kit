package sonar

import (
	"context"
	"time"
)

type DummyProvider struct{}

func (p *DummyProvider) Open(context.Context) chan Reading {
	c := make(chan Reading)

	go func() {
		for i := 1; i <= 10; i++ {
			c <- Reading{
				DistanceCM: 0,
				DurationUS: 0,
				Timestamp:  time.Now(),
			}
			time.Sleep(1000 * time.Millisecond)
		}
		close(c)
	}()

	return c
}

func (p *DummyProvider) Close(context.Context) error {
	return nil
}
