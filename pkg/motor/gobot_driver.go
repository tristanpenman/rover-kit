package motor

import (
	"context"
)

type GobotDriver struct{}

func (GobotDriver) Forwards(context.Context) error {
	return nil
}

func (GobotDriver) Backwards(context.Context) error {
	return nil
}

func (GobotDriver) SpinCW(context.Context) error {
	return nil
}

func (GobotDriver) SpinCCW(context.Context) error {
	return nil
}

func (GobotDriver) Stop(context.Context) error {
	return nil
}

func (GobotDriver) Throttle(_ context.Context, value float64) (bool, error) {
	return false, nil
}
