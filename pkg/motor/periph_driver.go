package motor

import (
	"context"
)

type PeriphDriver struct{}

func (PeriphDriver) Forwards(context.Context) error {
	return nil
}

func (PeriphDriver) Backwards(context.Context) error {
	return nil
}

func (PeriphDriver) SpinCW(context.Context) error {
	return nil
}

func (PeriphDriver) SpinCCW(context.Context) error {
	return nil
}

func (PeriphDriver) Stop(context.Context) error {
	return nil
}

func (PeriphDriver) Throttle(_ context.Context, value float64) (bool, error) {
	return false, nil
}
