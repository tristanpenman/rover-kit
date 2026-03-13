package motor

import (
	"context"
)

type AdafruitDriver struct{}

func (AdafruitDriver) Forwards(context.Context) error {
	return nil
}

func (AdafruitDriver) Backwards(context.Context) error {
	return nil
}

func (AdafruitDriver) SpinCW(context.Context) error {
	return nil
}

func (AdafruitDriver) SpinCCW(context.Context) error {
	return nil
}

func (AdafruitDriver) Stop(context.Context) error {
	return nil
}

func (AdafruitDriver) Throttle(_ context.Context, value float64) (bool, error) {
	return false, nil
}
