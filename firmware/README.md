# Firmware

This directory contains code intended to be run on STM32 microcontrollers.

## Examples

### `hello`

A TinyGo "Hello World" example that communicates over UART (`115200` baud). It prints a hello line once per second and echoes bytes sent from the host.

### `sonar`

Streams framed sonar samples over UART using the `pkg/uart` protocol.

## Prerequisites

Download and install [STM32CubeProgrammer](https://www.st.com/en/development-tools/stm32cubeprog.html).

## Pinouts

![STM32 Discovery F3 / F4 Pinout Differences](../reference/STM32-Discovery-F3-F4-Pinout-Differences.jpg)

## References

* [STM32 Discovery-F3 and Discovery-F4 Differences](https://kornakprotoblog.blogspot.com/2012/10/stm32-discovery-f3-and-discovery-f4.html)
