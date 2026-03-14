package motor

import (
	"context"
	"fmt"
	"math"
	"sync"

	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	defaultThreshold   = 0.5
	defaultMotorHatI2C = 0x60
)

type GobotDriver struct {
	mu        sync.Mutex
	adaptor   *raspi.Adaptor
	hat       *i2c.AdafruitMotorHatDriver
	threshold float64
}

func NewGobotDriver() (*GobotDriver, error) {
	return NewGobotDriverWithThreshold(defaultThreshold)
}

func NewGobotDriverWithThreshold(threshold float64) (*GobotDriver, error) {
	adaptor := raspi.NewAdaptor()
	if err := adaptor.Connect(); err != nil {
		return nil, fmt.Errorf("connect raspi adaptor: %w", err)
	}

	hat := i2c.NewAdafruitMotorHatDriver(
		adaptor,
		i2c.WithBus(1),
		i2c.WithAddress(defaultMotorHatI2C),
	)
	if err := hat.Start(); err != nil {
		_ = adaptor.Finalize()
		return nil, fmt.Errorf("start adafruit motor hat driver: %w", err)
	}

	driver := &GobotDriver{
		adaptor:   adaptor,
		hat:       hat,
		threshold: threshold,
	}
	if err := driver.Stop(context.Background()); err != nil {
		_ = hat.Halt()
		_ = adaptor.Finalize()
		return nil, fmt.Errorf("stop motors during startup: %w", err)
	}

	return driver, nil
}

func (d *GobotDriver) setMotor(motor int, throttle float64) error {
	t := math.Max(-1, math.Min(1, throttle))
	speed := int32(math.Round(math.Abs(t) * 255))

	if err := d.hat.SetDCMotorSpeed(motor, speed); err != nil {
		return err
	}

	switch {
	case t > 0:
		return d.hat.RunDCMotor(motor, i2c.AdafruitForward)
	case t < 0:
		return d.hat.RunDCMotor(motor, i2c.AdafruitBackward)
	default:
		return d.hat.RunDCMotor(motor, i2c.AdafruitRelease)
	}
}

func (d *GobotDriver) Stop(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, 0); err != nil {
			return fmt.Errorf("stop motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *GobotDriver) Forwards(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	values := [4]float64{-1, 1, -1, 1}
	for idx, value := range values {
		motor := idx + 1
		if err := d.setMotor(motor, value); err != nil {
			return fmt.Errorf("forwards motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *GobotDriver) Backwards(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	values := [4]float64{1, -1, 1, -1}
	for idx, value := range values {
		motor := idx + 1
		if err := d.setMotor(motor, value); err != nil {
			return fmt.Errorf("backwards motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *GobotDriver) SpinCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, 1); err != nil {
			return fmt.Errorf("spin_cw motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *GobotDriver) SpinCCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, -1); err != nil {
			return fmt.Errorf("spin_ccw motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *GobotDriver) Throttle(_ context.Context, value float64) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if math.Abs(value) <= d.threshold {
		for motor := 1; motor <= 4; motor++ {
			if err := d.setMotor(motor, 0); err != nil {
				return false, fmt.Errorf("threshold stop motor %d: %w", motor, err)
			}
		}
		return false, nil
	}

	values := [4]float64{value, -value, value, -value}
	for idx, throttle := range values {
		motor := idx + 1
		if err := d.setMotor(motor, throttle); err != nil {
			return false, fmt.Errorf("throttle motor %d: %w", motor, err)
		}
	}

	return true, nil
}

func (d *GobotDriver) Close(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.hat.Halt(); err != nil {
		return fmt.Errorf("halt adafruit motor hat driver: %w", err)
	}
	if err := d.adaptor.Finalize(); err != nil {
		return fmt.Errorf("finalize raspi adaptor: %w", err)
	}
	return nil
}
