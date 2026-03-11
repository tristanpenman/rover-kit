package common

import (
	"context"
	"log"
)

type DummyDriver struct{}

func (DummyDriver) Forwards(context.Context) error {
	log.Println("motor: forwards")
	return nil
}

func (DummyDriver) Backwards(context.Context) error {
	log.Println("motor: backwards")
	return nil
}

func (DummyDriver) SpinCW(context.Context) error {
	log.Println("motor: spin_cw")
	return nil
}

func (DummyDriver) SpinCCW(context.Context) error {
	log.Println("motor: spin_ccw")
	return nil
}

func (DummyDriver) Stop(context.Context) error {
	log.Println("motor: stop")
	return nil
}

func (DummyDriver) Throttle(_ context.Context, value float64) (bool, error) {
	active := value > 0.5 || value < -0.5
	log.Printf("motor: throttle value=%0.2f active=%v", value, active)
	return active, nil
}
