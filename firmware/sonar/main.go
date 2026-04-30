//go:build tinygo

package main

import (
	"machine"
	"time"

	"rover-kit/pkg/uart"
)

const (
	sensorCount         = 1
	streamInterval      = 100 * time.Millisecond
	sensorErrorSentinel = 0xFFFF
)

var (
	hostUART = machine.DefaultUART
)

type sensor struct {
	trigger machine.Pin
	echo    machine.Pin
}

var sensors = [sensorCount]sensor{
	{
		trigger: machine.PA0,
		echo: machine.PA1,
	},
}

func main() {
	configureIO()

	streaming := true
	lastSample := time.Now()
	boot := time.Now()

	for {
		for hostUART.Buffered() > 0 {
			cmd, err := hostUART.ReadByte()
			if err != nil {
				continue
			}
			switch cmd {
			case 'S':
				// start stream
				streaming = true
			case 'P':
				// pause stream
				streaming = false
			case 'O':
				// one-shot sample
				emitSample(uint32(time.Since(boot) / time.Millisecond))
			}
		}

		if streaming && time.Since(lastSample) >= streamInterval {
			emitSample(uint32(time.Since(boot) / time.Millisecond))
			lastSample = time.Now()
		}

		time.Sleep(2 * time.Millisecond)
	}
}

func configureIO() {
	hostUART.Configure(machine.UARTConfig{
		BaudRate: 115200,
	})

	for _, s := range sensors {
		s.trigger.Configure(machine.PinConfig{Mode: machine.PinOutput})
		s.trigger.Low()
		s.echo.Configure(machine.PinConfig{Mode: machine.PinInput})
	}
}

func emitSample(tsMS uint32) {
	readings := make([]uint16, sensorCount)
	for i, s := range sensors {
		distanceMM, ok := sampleDistanceMM(s)
		if !ok {
			readings[i] = sensorErrorSentinel
			continue
		}
		readings[i] = distanceMM
	}

	payload := uart.SampleV1{
		TimestampMS:  tsMS,
		DistanceUnit: uart.DistanceUnitMillimeters,
		Readings:     readings,
	}.MarshalPayload()

	frame := uart.EncodeFrame(uart.Version1, payload)
	_, _ = hostUART.Write(frame)
}

func sampleDistanceMM(s sensor) (uint16, bool) {
	// TODO: replace this stub with real HC-SR04 pulse timing.
	// Trigger pin for ~10us, then time ECHO pulse width and convert to distance.
	s.trigger.High()
	time.Sleep(10 * time.Microsecond)
	s.trigger.Low()

	return 250, true
}
