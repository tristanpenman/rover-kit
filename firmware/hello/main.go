//go:build tinygo

package main

import (
	"machine"
	"time"
)

var hostUART = machine.DefaultUART

func main() {
	hostUART.Configure(machine.UARTConfig{
		BaudRate: 115200,
		TX:       machine.PA9,
    RX:       machine.PA10,
	})

	_, _ = hostUART.Write([]byte("hello example booted\r\n"))

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		for hostUART.Buffered() > 0 {
			b, err := hostUART.ReadByte()
			if err != nil {
				continue
			}
			_, _ = hostUART.Write([]byte{'>', ' ', b, '\r', '\n'})
		}

		select {
		case <-ticker.C:
			_, _ = hostUART.Write([]byte("Hello, World from STM32 over UART!\r\n"))
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
}
