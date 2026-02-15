package driver

import "context"

type MotorDriver interface {
	Forwards(ctx context.Context) error
	Backwards(ctx context.Context) error
	SpinCW(ctx context.Context) error
	SpinCCW(ctx context.Context) error
	Throttle(ctx context.Context, value float64) (bool, error)
	Stop(ctx context.Context) error
}
