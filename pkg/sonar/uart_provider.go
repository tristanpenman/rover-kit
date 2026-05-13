package sonar

import (
	"context"
	"log"
	"time"

	// internal
	"rover-kit/pkg/uart"

	// third-party
	"go.bug.st/serial"
)

type UartProvider struct {
	port serial.Port
}

func NewUartProvider(portName string) (*UartProvider, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}

	return &UartProvider{
		port,
	}, nil
}

func (p *UartProvider) Open(context.Context) chan Reading {
	c := make(chan Reading)

	go func() {
		defer close(c)

		buff := make([]byte, 128)
		decoder := uart.NewDecoder(64)

		for {
			n, err := p.port.Read(buff)
			if err != nil {
				log.Printf("sonar uart read error: %v", err)
				return
			}
			if n == 0 {
				log.Println("sonar uart EOF reached")
				return
			}

			decoder.Push(buff[:n])
			for {
				frame, ok := decoder.NextFrame()
				if !ok {
					break
				}

				version, payload, err := uart.DecodeFrame(frame)
				if err != nil {
					log.Printf("sonar uart decode error: %v", err)
					continue
				}
				if version != uart.Version1 {
					log.Printf("sonar uart unsupported frame version: %d", version)
					continue
				}

				sample, err := uart.ParsePayloadV1(payload)
				if err != nil {
					log.Printf("sonar uart invalid v1 payload: %v", err)
					continue
				}

				for _, reading := range sampleToReadings(sample) {
					c <- reading
				}
			}
		}
	}()

	return c
}

func (p *UartProvider) Close(context.Context) error {
	err := p.port.Close()
	if err != nil {
		return err
	}
	return nil
}

func sampleToReadings(sample uart.SampleV1) []Reading {
	readings := make([]Reading, 0, len(sample.Readings))
	now := time.Now()

	for _, distance := range sample.Readings {
		if distance == 0xFFFF {
			continue
		}

		distanceCM := float64(distance)
		if sample.DistanceUnit == uart.DistanceUnitMillimeters {
			distanceCM = distanceCM / 10
		}

		readings = append(readings, Reading{
			SonarIndex: 0,
			DistanceCM: distanceCM,
			DurationUS: 0,
			Timestamp:  now,
		})
	}

	return readings
}
