# Rover Kit

Adventures in building a toy rover that can respond to commands over Wi-Fi and send back readings from ultrasonic distance sensors.

![Sensor mounted, not quite wired up...](./photos/05-sensors-mounted.jpeg)

## Inspiration

This project is inspired by a [blog post by Mat Kelcey](https://matpalm.com/blog/drivebot/) about building a rover and training it to move around autonomously.

I've used the same basic parts, but taken the project in a different direction. My focus is primarily hardware hacking.

## Parts

The rover is based on the following parts:

* [Whippersnapper Runt Rover](https://www.servocity.com/whippersnapper-runt-rover)
* [Raspberry Pi Zero W](https://www.raspberrypi.com/products/raspberry-pi-zero-w)
* [Adafruit DC & Stepper Motor HAT](https://www.adafruit.com/product/2348)
* [HC-SR04 Ultrasonic Distance Sensor](https://www.sparkfun.com/products/15569) (x4)

This is all wired up with an assortment of resistors, jumper wires, and breadboards.

Power to the Motor HAT is provided by a 12V battery pack. Power to the Raspberry Pi is provided by a portable USB power supply.

### Prototype

When I first started this project, I was using a regular Raspberry Pi 3 with components connected via a breadboard:

![Early prototype](./photos/00-early-prototype.jpeg)

I eventually switched to using a Raspberry Pi Zero W, so that power usage and space requirements would be reduced:

![Switching to Raspberry Pi Zero W](./photos/01-switching-to-pi-zero.jpeg)

Next, I eliminated the ugly breadboard by soldering my own sonar sensor interface. This was very slow because I'm new to soldering:

![Half way through sensor interface board](./photos/03-half-way.jpeg)

Everything looked much neater with the new interface board. The next step was to figure out wiring:

![Figuring out wiring after soldering was complete](./photos/04-figuring-out-wiring.jpeg)

Other photos can be found [here](./photos).

## Source

There are two ways to run this project: _Python_ and _Go_.

### Python

My original implementation was written in Python. Instructions can be found in the [python](./python/) directory.

### Go

The project has since been migrated to Go.

## License

This code is licensed under the MIT License.

See the [LICENSE](./LICENSE) file for more information.
