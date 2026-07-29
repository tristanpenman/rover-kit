# Issues

These issues were identified while reviewing the repository against [`notes/conventions.md`](notes/conventions.md).

## High priority

### Put motors into a safe state during shutdown

`PeriphDriver.Close` closes the I2C bus without stopping the motors. `GobotDriver.Close` calls `Halt` without first releasing the motors; the Motor HAT `Halt` implementation in the version used by this repository is a no-op.

Affected code:

- [`pkg/motor/periph_driver.go`](pkg/motor/periph_driver.go)
- [`pkg/motor/gobot_driver.go`](pkg/motor/gobot_driver.go)

Expected outcome:

- Stop or release every motor before closing its bus, driver, or adaptor.
- Attempt remaining cleanup even if stopping a motor or halting the driver fails.
- Add tests around cleanup ordering using narrow hardware seams or fakes.

### Make sonar providers cancellation-safe

`DummyProvider.Open` ignores its context, uses an uncancellable channel send and sleep, and never closes its channel. `UartProvider.Open` also ignores its context and can remain blocked while publishing a reading after the serial port is closed.

Affected code:

- [`pkg/sonar/dummy_provider.go`](pkg/sonar/dummy_provider.go)
- [`pkg/sonar/uart_provider.go`](pkg/sonar/uart_provider.go)

Expected outcome:

- Observe cancellation while sending and waiting between samples.
- Close provider-owned output channels when their workers finish.
- Close the serial port to unblock a pending `Read`.
- Ensure shutdown cannot be prevented by a stalled consumer.

### Move provider loops out of MQTT connection callbacks

The sonar and camera commands open their provider and consume its channel inside `OnConnect`. These callbacks remain active for the lifetime of the provider and may start another reader if MQTT reconnects.

Affected code:

- [`cmd/sonar-reader/main.go`](cmd/sonar-reader/main.go)
- [`cmd/camera-reader/main.go`](cmd/camera-reader/main.go)

Expected outcome:

- Start each provider exactly once outside `OnConnect`.
- Keep `OnConnect` limited to short connection or subscription work.
- Publish through connection state that remains safe across reconnects.
- Add a test demonstrating that reconnecting does not create duplicate readers.

### Preserve split UART synchronization markers

The UART decoder discards its complete buffer when it cannot find `0xAA 0x55`. If one read ends in `0xAA` and the next begins with `0x55`, the partial synchronization marker is lost.

Affected code:

- [`pkg/uart/protocol.go`](pkg/uart/protocol.go)

Expected outcome:

- Retain a trailing `0xAA` while waiting for more bytes.
- Add protocol tests for fragmented markers, fragmented frames, coalesced frames, noise, bad lengths, bad CRCs, and resynchronization.

## Medium priority

### Serialize writes to each WebSocket connection

Sonar broadcasts, camera broadcasts, and the connection read loop may call `WriteMessage` concurrently on the same WebSocket. Gorilla WebSocket supports one concurrent writer per connection.

Affected code:

- [`cmd/web-bridge/main.go`](cmd/web-bridge/main.go)

Expected outcome:

- Give each client a write lock or a single writer goroutine.
- Keep the global client collection lock out of network I/O.
- Test concurrent broadcasts and command responses with the race detector.

### Validate throttle values at protocol boundaries

Motor control and the WebSocket bridge accept throttle values outside the normalized range and rely on hardware drivers to clamp them.

Affected code:

- [`cmd/motor-control/main.go`](cmd/motor-control/main.go)
- [`cmd/web-bridge/main.go`](cmd/web-bridge/main.go)

Expected outcome:

- Reject throttle values outside `[-1, 1]` before publishing commands or touching hardware.
- Share validation so both protocol boundaries behave consistently.
- Add boundary and out-of-range tests.

### Validate UART payload and sensor counts before encoding

`SampleV1.MarshalPayload` casts the sensor count to one byte, and `EncodeFrame` casts the payload length to one byte. Oversized inputs therefore produce malformed frames. `ParsePayloadV1` also accepts unknown distance units.

Affected code:

- [`pkg/uart/protocol.go`](pkg/uart/protocol.go)

Expected outcome:

- Make encoding fail when the sensor count or payload cannot be represented on the wire.
- Reject unsupported distance units while parsing version 1 payloads.
- Add maximum, oversized, and unsupported-unit tests.

### Restore host test compatibility for the serial integration

`go test ./...` does not compile `pkg/sonar` or `cmd/sonar-reader` on Darwin/arm64 because `go.bug.st/serial` v1.1.1 lacks the required platform implementation.

Affected code:

- [`go.mod`](go.mod)
- [`go.sum`](go.sum)

Expected outcome:

- Use a serial dependency version or integration boundary that supports the development and CI platforms.
- Confirm that `go test ./...` passes without physical serial hardware.

## Low priority

### Correct and improve hardware error context

Some Periph GPIO setup operations return raw errors. Several motor methods label failures with the wrong operation, such as reporting `spin_cw` from `Forwards`.

Affected code:

- [`pkg/sonar/periph_provider.go`](pkg/sonar/periph_provider.go)
- [`pkg/motor/periph_driver.go`](pkg/motor/periph_driver.go)
- [`pkg/motor/gobot_driver.go`](pkg/motor/gobot_driver.go)

Expected outcome:

- Wrap GPIO configuration failures with the pin and requested operation.
- Report the actual motor operation and motor index from every driver method.

### Document exported package APIs

Many exported types, constants, interfaces, constructors, and protocol functions do not have Go documentation comments.

Affected packages:

- [`pkg/common`](pkg/common)
- [`pkg/motor`](pkg/motor)
- [`pkg/sonar`](pkg/sonar)
- [`pkg/camera`](pkg/camera)
- [`pkg/uart`](pkg/uart)

Expected outcome:

- Add concise comments for exported APIs, starting with the public protocol and hardware interfaces.

## Verification baseline

The review established the following baseline:

- `gofmt -l` reports no files.
- All four commands cross-compile for Linux ARMv6.
- Both TinyGo firmware targets compile for `stm32f4disco`.
- `go test ./...` passes for camera, motor control, web bridge, common, and motor packages.
- `go test ./...` fails to compile the sonar packages on Darwin/arm64 because of `go.bug.st/serial` v1.1.1.
- No physical hardware tests were performed.
