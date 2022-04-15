#!/usr/bin/env python3

import time
from adafruit_motorkit import MotorKit


kit = MotorKit()


def test_motors():
    kit.motor1.throttle = 1.0
    time.sleep(0.5)
    kit.motor1.throttle = 0

    time.sleep(0.5)

    kit.motor2.throttle = 1.0
    time.sleep(0.5)
    kit.motor2.throttle = 0

    time.sleep(0.5)

    kit.motor3.throttle = 1.0
    time.sleep(0.5)
    kit.motor3.throttle = 0

    time.sleep(0.5)

    kit.motor4.throttle = 1.0
    time.sleep(0.5)
    kit.motor4.throttle = 0


if __name__ == "__main__":
    test_motors()
