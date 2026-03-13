# Motors

Because the Adafruit Motor HAT is an I2C device (PCA9685 + direction-control GPIO expander), the existing `pkg/motor.MotorDriver` interface is a good seam for swapping runtime implementations.

## Option A: `periph.io` (low-level, explicit control)

Use this when you want direct control over I2C transactions and minimal framework overhead.

```go
package motor

import (
    "context"
    "fmt"

    "periph.io/x/conn/v3/i2c"
    "periph.io/x/host/v3"
)

type PeriphHATDriver struct {
    bus  i2c.BusCloser
    addr uint16
}

func NewPeriphHATDriver(addr uint16) (*PeriphHATDriver, error) {
    if _, err := host.Init(); err != nil {
        return nil, fmt.Errorf("init periph host: %w", err)
    }

    bus, err := i2c.Open("") // default Pi I2C bus (usually /dev/i2c-1)
    if err != nil {
        return nil, fmt.Errorf("open i2c bus: %w", err)
    }

    d := &PeriphHATDriver{bus: bus, addr: addr}
    if err := d.initPWM(); err != nil {
        _ = bus.Close()
        return nil, err
    }
    return d, nil
}

func (d *PeriphHATDriver) Forwards(ctx context.Context) error {
    _ = ctx
    return d.setMotors(1.0, 1.0)
}

func (d *PeriphHATDriver) Backwards(ctx context.Context) error {
    _ = ctx
    return d.setMotors(-1.0, -1.0)
}

func (d *PeriphHATDriver) SpinCW(ctx context.Context) error {
    _ = ctx
    return d.setMotors(1.0, -1.0)
}

func (d *PeriphHATDriver) SpinCCW(ctx context.Context) error {
    _ = ctx
    return d.setMotors(-1.0, 1.0)
}

func (d *PeriphHATDriver) Stop(ctx context.Context) error {
    _ = ctx
    return d.setMotors(0, 0)
}

func (d *PeriphHATDriver) Throttle(ctx context.Context, value float64) (bool, error) {
    _ = ctx
    if err := d.setMotors(value, value); err != nil {
        return false, err
    }
    return value > 0.05 || value < -0.05, nil
}

// initPWM/setMotors/writeReg would encode the Motor HAT register map.
```

Wire it into `cmd/motor-control/main.go` by replacing:

```go
dummyDriver := common.DummyDriver{}
```

with:

```go
motorDriver, err := motor.NewPeriphHATDriver(0x60)
if err != nil {
    log.Fatalf("create periph motor hat driver: %v", err)
}
defer motorDriver.Close()
```

## Option B: `gobot` (robotics framework)

Use this when you want to share a single framework across sensors, scheduling, and hardware orchestration.

```go
package motor

import (
    "context"

    "gobot.io/x/gobot/v2"
    "gobot.io/x/gobot/v2/platforms/raspi"
)

type GobotHATDriver struct {
    bot     *gobot.Robot
    adaptor *raspi.Adaptor
    hatAddr int
}

func NewGobotHATDriver(addr int) (*GobotHATDriver, error) {
    adaptor := raspi.NewAdaptor()

    work := func() {
        // initialize PCA9685 and motor channels over adaptor.GetConnection("i2c")
    }

    robot := gobot.NewRobot("motor-hat", []gobot.Connection{adaptor}, work)
    if err := robot.Start(false); err != nil {
        return nil, err
    }

    return &GobotHATDriver{bot: robot, adaptor: adaptor, hatAddr: addr}, nil
}

func (d *GobotHATDriver) Forwards(context.Context) error  { return d.setMotors(1, 1) }
func (d *GobotHATDriver) Backwards(context.Context) error { return d.setMotors(-1, -1) }
func (d *GobotHATDriver) SpinCW(context.Context) error    { return d.setMotors(1, -1) }
func (d *GobotHATDriver) SpinCCW(context.Context) error   { return d.setMotors(-1, 1) }
func (d *GobotHATDriver) Stop(context.Context) error      { return d.setMotors(0, 0) }
func (d *GobotHATDriver) Throttle(context.Context, v float64) (bool, error) {
    return v > 0.05 || v < -0.05, d.setMotors(v, v)
}
```

## Option C: Go standard library + Linux syscalls (fewest dependencies)

Use this when you want no third-party hardware dependency and are comfortable owning Linux `ioctl` + I2C details.

```go
package motor

import (
    "golang.org/x/sys/unix"
    "os"
)

const i2cSlave = 0x0703

type StdlibHATDriver struct {
    f    *os.File
    addr int
}

func NewStdlibHATDriver(bus string, addr int) (*StdlibHATDriver, error) {
    f, err := os.OpenFile(bus, os.O_RDWR, 0)
    if err != nil {
        return nil, err
    }
    if err := unix.IoctlSetInt(int(f.Fd()), i2cSlave, addr); err != nil {
        _ = f.Close()
        return nil, err
    }
    return &StdlibHATDriver{f: f, addr: addr}, nil
}

func (d *StdlibHATDriver) writeReg(reg, value byte) error {
    _, err := d.f.Write([]byte{reg, value})
    return err
}
```

## Recommendation for this repo

- If motor control remains on Raspberry Pi only, `periph.io` is usually the best fit: small API surface and straightforward mapping to the HAT register model.
- If you expect to manage more robotic components with a common event loop, choose `gobot`.
- The syscall approach is viable but higher maintenance; use it only when minimizing dependencies is the top priority.
