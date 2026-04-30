package sonar

import (
	"context"

	// third-party
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type PeriphProvider struct {
	trig gpio.PinOut
	echo gpio.PinIn
	// timeout time.Duration
}

func NewPeriphProvider() (*PeriphProvider, error) {
	trig := gpioreg.ByName("GPIO23")
	echo := gpioreg.ByName("GPIO24")

	return &PeriphProvider{
		trig: trig,
		echo: echo,
	}, nil
}

func (p *PeriphProvider) Open(context.Context) chan Reading {
	return make(chan Reading)
}

func (p *PeriphProvider) Close(context.Context) error {
	return nil
}
