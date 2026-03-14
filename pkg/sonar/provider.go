package sonar

import "context"

type Provider interface {
	Close(ctx context.Context) error
}
