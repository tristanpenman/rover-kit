package sonar

import (
	"context"
	"fmt"
	"log"
	"time"

	// third-party
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	defaultTriggerPin = "GPIO18"
	defaultEchoPin    = "GPIO24"
	triggerPulse      = 10 * time.Microsecond
	sampleInterval    = 1 * time.Second
	echoTimeout       = 30 * time.Millisecond
	soundSpeedCMPerS  = 34300.0
)

type PeriphProvider struct {
	trig gpio.PinOut
	echo gpio.PinIn
}

func NewPeriphProvider() (*PeriphProvider, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("initialize periph host: %w", err)
	}

	trig := gpioreg.ByName(defaultTriggerPin)
	if trig == nil {
		return nil, fmt.Errorf("trigger pin %s not found", defaultTriggerPin)
	}

	echo := gpioreg.ByName(defaultEchoPin)
	if echo == nil {
		return nil, fmt.Errorf("echo pin %s not found", defaultEchoPin)
	}

	// no pull on echo pin
	err := echo.In(gpio.Float, gpio.NoEdge)
	if err != nil {
		return nil, err
	}

	// default trigger pin to low
	err = trig.Out(gpio.Low)
	if err != nil {
		return nil, err
	}

	return &PeriphProvider{
		trig: trig,
		echo: echo,
	}, nil
}

func (p *PeriphProvider) Open(context.Context) chan Reading {
	c := make(chan Reading)

	go func() {
		defer close(c)

		for {
			// set trigger pin high
			err := p.trig.Out(gpio.High)
			if err != nil {
				log.Printf("error setting trig high: %v", err)
				return
			}

			// sleep for `triggerPulse` nanoseconds
			time.Sleep(triggerPulse)

			// set trigger low
			err = p.trig.Out(gpio.Low)
			if err != nil {
				log.Printf("error setting trig low: %v", err)
				return
			}

			start := time.Now()

			// wait for echo high
			for p.echo.Read() != gpio.High {
				if time.Since(start) > echoTimeout {
					log.Printf("timed out waiting for echo high")
					return
				}
			}

			start = time.Now()

			// wait for echo low
			for p.echo.Read() != gpio.Low {
				if time.Since(start) > echoTimeout {
					log.Printf("timed out waiting for echo low")
					return
				}
			}

			end := time.Now()

			duration := end.Sub(start)

			c <- Reading{
				DistanceCM: duration.Seconds() * soundSpeedCMPerS / 2,
				DurationUS: float64(duration.Microseconds()),
				Timestamp:  time.Now(),
			}

			time.Sleep(sampleInterval)
		}
	}()

	return c
}

func (p *PeriphProvider) Close(context.Context) error {
	return nil
}
