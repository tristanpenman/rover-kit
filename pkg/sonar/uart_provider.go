package sonar

import (
	"context"
	"log"

	// internal
	"rover-kit/pkg/common"

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
		var lb common.LineBuffer

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

			lines := lb.Append(buff[:n])
			for _, line := range lines {
				// TODO: parse uart frames into Readings via pkg/uart Decoder
				log.Printf("sonar uart line: %s", line)
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
