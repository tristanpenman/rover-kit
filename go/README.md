# Go Workspace

This directory introduces a modular Go layout for motor control and sonar readings.

This consists of three commands:

- `cmd/motor-control` - Subscribes to typed motor commands and invokes a `MotorDriver` implementation
- `cmd/sonar-reader` - Consumes a `SonarProvider` implementation and publishes distance events
- `cmd/web-bridge` - HTTP/WebSocket bridge from browser clients into broker topics

Shared packages include:

- `pkg/messages` - Message types, strongly-typed command/event models, and broker abstractions
- `pkg/motor` - Defines a `MotorDriver` interface and local implementation using GPIO
- `pkg/sonar` - Defines a `SonarProvider` interface and local implementation using GPIO

Communication between components will be handled by MQTT.

## Local MQTT broker

The [mqtt](mqtt) directory contains configuration for a [Mosquitto](https://mosquitto.org/) message broker. Mosquitto is a popular and light-weight message broker that implements the MQTT protocol.

To start the local MQTT broker you'll need to use Docker Compose.

From the `go` directory:

```bash
docker compose up -d mqtt
```

To stop the broker:

```bash
docker compose down
```

### Injecting Commands

Publish a command to MQTT, you can use `mosquitto_pub`:

```bash
mosquitto_pub -h localhost -p 1883 -t rover/motor/command -m '{"type":"forwards"}'
```

To use `mosquitto_pub` on macOS, install `mosquitto` from Homebrew:

```bash
brew install mosquitto
```

Alternatively, you can run `mosquitto_pub` commands from within the `mqtt` Docker container created above:

```bash
docker compose exec mqtt mosquitto_pub -t rover/motor/command -m '{"type":"throttle","value":0.75}'
```
