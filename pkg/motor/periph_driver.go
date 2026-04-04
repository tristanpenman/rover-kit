package motor

import (
	"context"
	"fmt"
	"math"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/pca9685"
	"periph.io/x/host/v3"
)

type periphMotorChannel struct {
	pwm int
	in1 int
	in2 int
}

var periphMotorChannels = [4]periphMotorChannel{
	{pwm: 8, in1: 10, in2: 9},
	{pwm: 13, in1: 11, in2: 12},
	{pwm: 2, in1: 4, in2: 3},
	{pwm: 7, in1: 5, in2: 6},
}

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

func (d *PeriphDriver) setChannel(channel int, on bool) error {
	if on {
		return d.pca.SetFullOn(channel)
	}
	return d.pca.SetFullOff(channel)
}

func (d *PeriphDriver) setMotor(motor int, throttle float64) error {
	if motor < 1 || motor > len(periphMotorChannels) {
		return fmt.Errorf("unsupported motor index %d", motor)
	}
	ch := periphMotorChannels[motor-1]

	t := math.Max(-1, math.Min(1, throttle))
	duty := gpio.Duty(math.Round(math.Abs(t) * float64(gpio.DutyMax)))

	switch {
	case t > 0:
		if err := d.setChannel(ch.in1, true); err != nil {
			return fmt.Errorf("set in1 high: %w", err)
		}
		if err := d.setChannel(ch.in2, false); err != nil {
			return fmt.Errorf("set in2 low: %w", err)
		}
	case t < 0:
		if err := d.setChannel(ch.in1, false); err != nil {
			return fmt.Errorf("set in1 low: %w", err)
		}
		if err := d.setChannel(ch.in2, true); err != nil {
			return fmt.Errorf("set in2 high: %w", err)
		}
	default:
		if err := d.setChannel(ch.in1, false); err != nil {
			return fmt.Errorf("release in1: %w", err)
		}
		if err := d.setChannel(ch.in2, false); err != nil {
			return fmt.Errorf("release in2: %w", err)
		}
	}

	if duty == 0 {
		return d.pca.SetFullOff(ch.pwm)
	}
	return d.pca.SetPwm(ch.pwm, 0, duty)
}

func (d *PeriphDriver) Stop(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, 0); err != nil {
			return fmt.Errorf("stop motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *PeriphDriver) Forwards(context.Context) error {
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

func (d *PeriphDriver) Backwards(context.Context) error {
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

func (d *PeriphDriver) SpinCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, 1); err != nil {
			return fmt.Errorf("spin_cw motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *PeriphDriver) SpinCCW(context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for motor := 1; motor <= 4; motor++ {
		if err := d.setMotor(motor, -1); err != nil {
			return fmt.Errorf("spin_ccw motor %d: %w", motor, err)
		}
	}
	return nil
}

func (d *PeriphDriver) Throttle(_ context.Context, value float64) (bool, error) {
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
