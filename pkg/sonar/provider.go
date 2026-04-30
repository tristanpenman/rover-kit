package sonar

import (
	"context"
	"time"
)

type Reading struct {
	DistanceCM float64
	DurationUS float64
	Timestamp  time.Time
}

type Provider interface {
	Open(ctx context.Context) chan Reading

	Close(ctx context.Context) error
}
