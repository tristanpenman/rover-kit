package motor

import (
	"context"
	"fmt"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/pca9685"
	"periph.io/x/host/v3"
)

type PeriphDriver struct {
	mu        sync.Mutex
	bus       i2c.BusCloser
	pca       *pca9685.Dev
	threshold float64
}

const (
	defaultPeriphThreshold   = 0.5
	defaultPeriphMotorHatI2C = 0x60
)

func NewPeriphDriver() (*PeriphDriver, error) {
	return NewPeriphDriverWithThreshold(defaultPeriphThreshold)
}

func NewPeriphDriverWithThreshold(threshold float64) (*PeriphDriver, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("initialize periph host: %w", err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		return nil, fmt.Errorf("open i2c bus: %w", err)
	}

	dev, err := pca9685.NewI2C(bus, defaultPeriphMotorHatI2C)
	if err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("initialize pca9685 on i2c address 0x%X: %w", defaultPeriphMotorHatI2C, err)
	}
	if err := dev.SetPwmFreq(1600); err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("set pca9685 pwm frequency: %w", err)
	}

	driver := &PeriphDriver{bus: bus, pca: dev, threshold: threshold}
	if err := driver.Stop(context.Background()); err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("stop motors during startup: %w", err)
	}

	return driver, nil
}

func (d *PeriphDriver) Forwards(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *PeriphDriver) Backwards(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *PeriphDriver) SpinCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *PeriphDriver) SpinCCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *PeriphDriver) Stop(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *PeriphDriver) Throttle(_ context.Context, value float64) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return false, nil
}

func (d *PeriphDriver) Close(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.bus != nil {
		if err := d.bus.Close(); err != nil {
			return fmt.Errorf("close i2c bus: %w", err)
		}
	}

	return nil
}
