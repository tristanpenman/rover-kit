# STM32

This document sketches how to port the Python `test_sensors.py` HC-SR04 logic to Embedded Go (TinyGo) for STM32 hardware.

## Goals

- Run sonar measurement timing directly on STM32 GPIO.
- Exchange command and telemetry messages over USB CDC serial.
- Keep a small, explicit protocol so a host app (Go/Python/ROS bridge) can control sampling.

## Mapping from Python to TinyGo

The Python flow is:

1. Raise trigger pin high for ~10us.
2. Wait for echo pin to go high (pulse start).
3. Wait for echo pin to go low (pulse end).
4. Compute distance from pulse width.

Equivalent TinyGo flow is nearly identical, with nanosecond/microsecond timing from `time.Now()` and tight polling loops on `machine.Pin.Get()`.

## Suggested USB Message Protocol

Use newline-delimited JSON over USB CDC.

Host -> STM32 commands:

```json
{"type":"ping"}
{"type":"set_interval_ms","value":100}
{"type":"sample_once"}
{"type":"start_stream"}
{"type":"stop_stream"}
```

STM32 -> Host events:

```json
{"type":"pong"}
{"type":"distance_cm","sensor":"front_left","value":37.2,"ok":true}
{"type":"distance_cm","sensor":"front_left","ok":false,"error":"echo_timeout"}
{"type":"status","streaming":true,"interval_ms":100}
```

## TinyGo Skeleton

```go
//go:build tinygo && stm32

package main

import (
    "encoding/json"
    "machine"
    "time"
)

type Command struct {
    Type  string `json:"type"`
    Value int    `json:"value,omitempty"`
}

type DistanceEvent struct {
    Type   string  `json:"type"`
    Sensor string  `json:"sensor"`
    Value  float32 `json:"value,omitempty"`
    OK     bool    `json:"ok"`
    Error  string  `json:"error,omitempty"`
}

var (
    trigger = machine.PA0
    echo    = machine.PA1

    // Depending on STM32 target this may be machine.Serial, machine.USBCDC, or machine.USB.
    usb = machine.Serial
)

func main() {
    trigger.Configure(machine.PinConfig{Mode: machine.PinOutput})
    echo.Configure(machine.PinConfig{Mode: machine.PinInput})

    usb.Configure(machine.UARTConfig{})

    streaming := false
    interval := 100 * time.Millisecond

    go readCommands(func(cmd Command) {
        switch cmd.Type {
        case "ping":
            writeJSON(map[string]any{"type": "pong"})
        case "set_interval_ms":
            if cmd.Value >= 20 {
                interval = time.Duration(cmd.Value) * time.Millisecond
            }
        case "sample_once":
            emitSample()
        case "start_stream":
            streaming = true
        case "stop_stream":
            streaming = false
        }
    })

    for {
        if streaming {
            emitSample()
            time.Sleep(interval)
        } else {
            time.Sleep(10 * time.Millisecond)
        }
    }
}

func emitSample() {
    cm, err := measureDistanceCM(30 * time.Millisecond)
    evt := DistanceEvent{Type: "distance_cm", Sensor: "front_left", OK: err == nil}
    if err != nil {
        evt.Error = err.Error()
    } else {
        evt.Value = cm
    }
    writeJSON(evt)
}

func measureDistanceCM(timeout time.Duration) (float32, error) {
    // 10us trigger pulse.
    trigger.High()
    time.Sleep(10 * time.Microsecond)
    trigger.Low()

    deadline := time.Now().Add(timeout)
    for !echo.Get() {
        if time.Now().After(deadline) {
            return 0, errEchoStartTimeout
        }
    }

    start := time.Now()
    deadline = time.Now().Add(timeout)
    for echo.Get() {
        if time.Now().After(deadline) {
            return 0, errEchoEndTimeout
        }
    }
    elapsed := time.Since(start)

    // Speed of sound ~34300 cm/s, divide by 2 for round trip.
    distanceCM := float32(elapsed.Seconds()*34300.0) / 2.0
    return distanceCM, nil
}

func readCommands(handle func(Command)) {
    // Implement as line reader from usb; unmarshal JSON per line.
    // Keep this strict to avoid memory growth.
}

func writeJSON(v any) {
    b, _ := json.Marshal(v)
    usb.Write(append(b, '\n'))
}
```

## Multi-Sensor Expansion

For 4x HC-SR04 sensors:

- Keep each sensor as `{triggerPin, echoPin, name}`.
- Trigger and read one sensor at a time to avoid acoustic cross-talk.
- Optional: median of N=3 readings per sensor to reduce spikes.
- Return one event per sensor or a batch event (tradeoff: simplicity vs throughput).

## Host-side USB bridge sketch (desktop Go)

- Open USB CDC serial port (e.g. `/dev/ttyACM0`).
- Send `start_stream` + `set_interval_ms`.
- Read line-delimited JSON events and forward to MQTT/WebSocket/ROS.

This keeps STM32 firmware small and deterministic while letting heavier processing run on host.

## STM32/TinyGo bring-up checklist

1. Confirm target board and the TinyGo target name.
2. Verify USB CDC endpoint support for that target.
3. Verify trigger/echo pins are 5V-safe (HC-SR04 echo often needs level shifting).
4. Calibrate timeout and minimum sample interval per sensor placement.
5. Add watchdog/reset strategy for production firmware.
