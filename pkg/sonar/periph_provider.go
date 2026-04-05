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

func (p *PeriphProvider) Sample(context.Context) (Reading, error) {
	panic("implement me")
}

func (p *PeriphProvider) Close(context.Context) error {
	return nil
}
