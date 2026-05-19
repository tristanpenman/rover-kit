package sonar

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	// third-party
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	triggerPulse     = 10 * time.Microsecond
	sampleInterval   = 1 * time.Second
	echoTimeout      = 30 * time.Millisecond
	soundSpeedCMPerS = 34300.0

	// pin configuration
	defaultTriggerPin = "GPIO18"
	defaultEchoPin    = "GPIO24"

	// pin configuration env vars
	sonarTriggerPin1Env = "SONAR_TRIGGER_PIN_1"
	sonarEchoPin1Env    = "SONAR_ECHO_PIN_1"
	sonarTriggerPin2Env = "SONAR_TRIGGER_PIN_2"
	sonarEchoPin2Env    = "SONAR_ECHO_PIN_2"
)

type PeriphProvider struct {
	sonars []periphSonar
}

type sonarPinConfig struct {
	index      int
	triggerPin string
	echoPin    string
}

type periphSonar struct {
	index int
	trig  gpio.PinOut
	echo  gpio.PinIn
}

func NewPeriphProvider() (*PeriphProvider, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("initialize periph host: %w", err)
	}

	configs, err := periphSonarPinConfigs()
	if err != nil {
		return nil, err
	}

	sonars := make([]periphSonar, 0, len(configs))
	for _, config := range configs {
		trig := gpioreg.ByName(config.triggerPin)
		if trig == nil {
			return nil, fmt.Errorf("trigger pin %s not found", config.triggerPin)
		}

		echo := gpioreg.ByName(config.echoPin)
		if echo == nil {
			return nil, fmt.Errorf("echo pin %s not found", config.echoPin)
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

		sonars = append(sonars, periphSonar{
			index: config.index,
			trig:  trig,
			echo:  echo,
		})
	}

	return &PeriphProvider{
		sonars: sonars,
	}, nil
}

func periphSonarPinConfigs() ([]sonarPinConfig, error) {
	configs := []sonarPinConfig{{
		index:      0,
		triggerPin: defaultTriggerPin,
		echoPin:    defaultEchoPin,
	}}

	// allow override of first sonar
	triggerPin1 := envTrimmed(sonarTriggerPin1Env)
	echoPin1 := envTrimmed(sonarEchoPin1Env)
	if triggerPin1 != "" || echoPin1 != "" {
		if triggerPin1 == "" || echoPin1 == "" {
			return nil, fmt.Errorf("%s and %s must both be set", sonarTriggerPin1Env, sonarEchoPin1Env)
		}
		configs[0].triggerPin = triggerPin1
		configs[0].echoPin = echoPin1
	}

	// optionally add a second sonar
	triggerPin2 := envTrimmed(sonarTriggerPin2Env)
	echoPin2 := envTrimmed(sonarEchoPin2Env)
	if triggerPin2 != "" || echoPin2 != "" {
		if triggerPin2 == "" || echoPin2 == "" {
			return nil, fmt.Errorf("%s and %s must both be set", sonarTriggerPin2Env, sonarEchoPin2Env)
		}
		configs = append(configs, sonarPinConfig{
			index:      1,
			triggerPin: triggerPin2,
			echoPin:    echoPin2,
		})
	}

	return configs, nil
}

func envTrimmed(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func (p *PeriphProvider) Open(ctx context.Context) chan Reading {
	c := make(chan Reading)

	var wg sync.WaitGroup
	wg.Add(len(p.sonars))
	for _, sonar := range p.sonars {
		go func() {
			defer wg.Done()
			readPeriphSonar(ctx, c, sonar)
		}()
	}

	go func() {
		defer close(c)
		wg.Wait()
	}()

	return c
}

func readPeriphSonar(ctx context.Context, c chan<- Reading, sonar periphSonar) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// set trigger pin high
		err := sonar.trig.Out(gpio.High)
		if err != nil {
			log.Printf("sonar %d error setting trig high: %v", sonar.index, err)
			return
		}

		// sleep for `triggerPulse` nanoseconds
		time.Sleep(triggerPulse)

		// set trigger low
		err = sonar.trig.Out(gpio.Low)
		if err != nil {
			log.Printf("sonar %d error setting trig low: %v", sonar.index, err)
			return
		}

		start := time.Now()

		// wait for echo high
		timedOut := false
		for sonar.echo.Read() != gpio.High {
			if time.Since(start) > echoTimeout {
				log.Printf("sonar %d timed out waiting for echo high", sonar.index)
				timedOut = true
				break
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		if timedOut {
			sleepOrDone(ctx, sampleInterval)
			continue
		}

		start = time.Now()

		// wait for echo low
		for sonar.echo.Read() != gpio.Low {
			if time.Since(start) > echoTimeout {
				log.Printf("sonar %d timed out waiting for echo low", sonar.index)
				timedOut = true
				break
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
		if timedOut {
			sleepOrDone(ctx, sampleInterval)
			continue
		}

		end := time.Now()

		duration := end.Sub(start)

		reading := Reading{
			SonarIndex: sonar.index,
			DistanceCM: duration.Seconds() * soundSpeedCMPerS / 2,
			DurationUS: float64(duration.Microseconds()),
			Timestamp:  time.Now(),
		}

		select {
		case c <- reading:
		case <-ctx.Done():
			return
		}

		sleepOrDone(ctx, sampleInterval)
	}
}

func sleepOrDone(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func (p *PeriphProvider) Close(context.Context) error {
	return nil
}
