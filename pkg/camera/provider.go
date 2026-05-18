package camera

import (
	"context"
	"time"
)

type Frame struct {
	Type        string    `json:"type"`
	ContentType string    `json:"content_type"`
	Data        string    `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
}

type Provider interface {
	Open(ctx context.Context) chan Frame

	Close(ctx context.Context) error
}
