# Conventions

These conventions complement standard Go guidance.

## Core Guidelines

### Structure

- Keep reusable logic in packages. Limit `cmd/<name>` to arguments, dependency construction, lifecycle, and UI wiring.
- Give cohesive, independently testable capabilities their own packages.
- Keep focused tests beside their package. Use a separate test package for black-box API tests. Top-level `test/` should only contain cross-package black-box coverage.
- Treat directories with their own manifest or instructions as separate components and follow their local guidance.

### Packaging

- Use `pkg/` for publicly importable code intended for consumers outside the repository. Treat its exported API as a compatibility commitment.
- Use `internal/` for code that supports the repository's `cmd/` programs and is not intended as a public library.
- Keep command-specific wiring in `cmd/`; move shared command implementation into focused `internal/` packages.
- For repositories that are pure libraries rather than collections of programs, prefer top-level packages over `pkg/` to keep import paths short.
- Choose package boundaries by responsibility and dependency direction, not merely by file size.

### Errors and logging

- Reusable packages return errors with concise operation context:

  ```go
  return nil, fmt.Errorf("service error: %w", err)
  ```

- Constructors clean up resources acquired before an error.
- Commands report actionable errors and clean up during shutdown. Panic only for unrecoverable process initialization.
- Use short structured log events with key/value context. Log full payloads only at debug level.
- Never log private keys, authentication secrets, salts, or similar data.
- Accept a logger from the owning layer where practical. Preserve the logging approach already established in an area.

### APIs and implementation

- Constructors normally use `New<Type>`. Preserve established names when compatibility matters; new fallible constructors should return an error.
- Export only what another package needs and document new exported APIs.
- Inject external effects through narrow function or interface seams.
- Return stable, presentation-ready collections when promised by the API. Preserve useful partial results alongside errors.
- Keep parsing and conversion deterministic and side-effect free.

### Formatting and tests

- Preserve a file's existing import grouping and run `gofmt` on changed Go files.
- Exclude generated code, reference documents, diagrams, and binary assets from broad formatting or replacement.
- Unit tests must not require devices, multicast networking, displays, hosted services, or cloud accounts.
- Cover success, failure, closure, timeout, fallback, deduplication, and stable ordering as relevant.
- Run focused package tests first, then `go test ./...` when integration dependencies permit.

## Idioms

These are primarily idioms that apply to this project. They might not apply elsewhere.

### Data and protocols

- Dispatch JSON commands by their small envelope before decoding the command-specific body. Reject unknown types and invalid numeric ranges.
- Treat UART as an arbitrary byte stream. A read may contain a partial frame, several frames, or noise.
- Publish only complete, CRC-valid UART frames. Bound payload lengths and recover from malformed data without losing the next valid frame.
- Preserve the protocol's little-endian encoding, unit field, wrapping `uint32` timestamp, and `0xFFFF` invalid-reading sentinel.

### Concurrency and lifecycle

- Derive command lifetimes from `signal.NotifyContext` and propagate cancellation to drivers, providers, and goroutines.
- A channel producer owns closing its channel; consumers range over it. Blocking sends and delays must also observe cancellation.
- Close a serial port or other connection to unblock a pending read during shutdown.
- Long-lived components expose `Close` or an equivalent lifecycle operation. Wait for required workers before closing their output channel.
- Protect actuator state and client collections with their owning mutex. Copy collections before slow network writes.
- Serialize compound motor and shared-bus operations. Do not let concurrent operations interleave direction and speed writes.
- Time-bounded waits distinguish cancellation, timeout, malformed input, and device closure.

### Configuration

- Use environment variables for broker URLs, topics, providers, drivers, device paths, and pins.
- Keep usable dummy or local defaults. Validate related settings together and name invalid variables in errors.
- Do not silently select the first serial port. Device numbering can change between boots and reconnects.
- Keep deployment-specific I2C buses, device addresses, and GPIO pins explicit or configurable.

### Web frontend

- Keep reusable protocol and message logic outside `cmd/web-bridge`.
- Keep `cmd/web-bridge/static` dependency-free unless a frontend toolchain provides a clear benefit.
- Preserve the build step that copies static assets beside the `web-bridge` binary.

### Documentation

- Change `notes/uart.md`, the host decoder, the firmware encoder, and their tests together when changing the UART protocol.

## Integrations

### MQTT

- Keep shared MQTT message types and command constants in `pkg/common`. Keep command-specific envelopes and responses private.
- Keep MQTT topic defaults with the command that owns them and allow deployment overrides through environment variables.
- Keep MQTT callbacks bounded. Reconnection must not start duplicate provider readers.

### TinyGo

- Guard firmware entry points with `//go:build tinygo`. Build them with the TinyGo targets, not ordinary `go build`:

  ```sh
  make tinygo-sonar
  make tinygo-hello
  # or: tinygo build -target="$TARGET" ./firmware/sonar
  ```

- Treat the target, UART, and named `machine` pins as hardware configuration. Verify them against the selected board; the debugger USB connection may not expose `machine.DefaultUART`.
- Keep shared code such as `pkg/uart` compatible with TinyGo. Avoid unsupported dependencies, unnecessary allocation, and unbounded sensor waits.
- Use the binary protocol in `notes/uart.md` for sonar telemetry. Emit `0xFFFF` for a failed reading and keep debug text off the framed stream.
- Follow `firmware/README.md` for the STM32F407 Discovery UART wiring.

### Serial

- Configure both ends as `115200 8N1` with no flow control:

  ```go
  mode := &serial.Mode{
      BaudRate: 115200,
      DataBits: 8,
      Parity:   serial.NoParity,
      StopBits: serial.OneStopBit,
  }
  port, err := serial.Open(portName, mode)
  ```

- Require an explicit port such as `/dev/ttyUSB0` or `/dev/ttyACM0`. Do not select the first enumerated device.
- Treat serial as a byte stream. Pass only `buf[:n]` to the bounded decoder; reads do not correspond to frames.
- Expect noise, partial frames, and resets. Resynchronize after invalid lengths or CRCs.
- Close the port to unblock `Read` during shutdown. Stop blocked channel sends on cancellation; if read timeouts are introduced, distinguish an idle timeout from EOF.

### Periph

- Initialize Periph before resolving pins or buses. Open and close I2C explicitly:

  ```go
  if _, err := host.Init(); err != nil {
      return nil, fmt.Errorf("initialize periph host: %w", err)
  }
  bus, err := i2creg.Open("")
  if err != nil {
      return nil, fmt.Errorf("open i2c bus: %w", err)
  }
  // Retain bus and close it on later failure and shutdown.
  ```

- Resolve pins through `gpioreg`, check for `nil`, and use SoC names such as `GPIO18`. Validate trigger/echo overrides as pairs and protect inputs from the HC-SR04's 5 V echo signal as documented in `notes/circuits.md`.
- Keep the I2C bus, PCA9685 address, PWM frequency, and Motor HAT channel map explicit. `i2creg.Open("")` means the host's default bus.
- Do not mix PWM scales: `pca9685.Dev.SetPwm` uses 12-bit counter values, while `gpio.PinIO.PWM` uses `gpio.DutyMax`.
- Check every GPIO and I2C operation. Include the motor or channel in errors and stop actuators during cleanup.
- Bound HC-SR04 edge waits. Linux GPIO polling is not hard real-time; use TinyGo/UART when timing matters.

### Gobot

- Follow Gobot's lifecycle in order:

  ```go
  adaptor := raspi.NewAdaptor()
  if err := adaptor.Connect(); err != nil {
      return nil, err
  }

  hat := i2c.NewAdafruitMotorHatDriver(
      adaptor,
      i2c.WithBus(1),
      i2c.WithAddress(0x60),
  )
  if err := hat.Start(); err != nil {
      _ = adaptor.Finalize()
      return nil, err
  }
  ```

- Stop or release motors before shutdown, then call `hat.Halt()` and `adaptor.Finalize()`. Unwind partial startup in the same reverse order.
- Keep bus `1`, Motor HAT address `0x60`, and the motor map explicit. Motor indexes are `0..3` and speeds are `0..255`; clamp normalized throttle and release the motor at zero.
- Serialize speed and direction writes because the I2C operations are not atomic.
- Keep Gobot lifecycle and types inside `pkg/motor` so all motor drivers remain interchangeable.
