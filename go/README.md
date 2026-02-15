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
