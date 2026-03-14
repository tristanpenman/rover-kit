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
	Sample(ctx context.Context) (Reading, error)

	Close(ctx context.Context) error
}
