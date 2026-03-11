# Python

This directory contains my original Python scripts for controlling the Rover.

## Dependencies

The scripts for controlling the rover are written in Python, and rely on some very useful libraries:

* [Adafruit_CircuitPython_MotorKit](https://github.com/adafruit/Adafruit_CircuitPython_MotorKit)
* [RPi.GPIO](https://pypi.org/project/RPi.GPIO)

Dependencies can be installed via `pip`:

```bash
pip install -r requirements.txt
```

## Control Script

The main control script ([rover_control.py](./rover_kit/rover_control.py)) does everything:

* Connects to Sonar sensor via GPIO
* Connects to Motor HAT via Adafruit MotorKit
* Starts a WebSocket server
* Starts a simple HTTP server to serve static content

## Tests

There are also a couple of test scripts:

* [test_motors.py](./rover_kit/test_motors.py) - Starts and stops each motor in turn
* [test_sensors.py](./rover_kit/test_sensors.py) - Outputs a series of ultrasonic sensor readings

## Frontend

The frontend is vanilla HTML and JavaScript. It currently provides the following commands:

* Forwards
* Backwards
* Spin Clockwise
* Spin Counter-clockwise
