package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// internal
	"rover-kit/pkg/common"
	"rover-kit/pkg/motor"

	// external
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL = "tcp://localhost:1883"
	defaultDriver    = "dummy"
	defaultTopic     = "rover/motor/command"
)

type commandEnvelope struct {
	Type common.CommandType `json:"type"`
}

func handleMotorCommand(ctx context.Context, driver motor.Driver, payload []byte) error {
	var envelope commandEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return fmt.Errorf("decode command envelope: %w", err)
	}

	switch envelope.Type {
	case common.CommandForwards:
		var cmd common.ForwardsCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode forwards command: %w", err)
		}
		return driver.Forwards(ctx)
	case common.CommandBackwards:
		var cmd common.BackwardsCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode backwards command: %w", err)
		}
		return driver.Backwards(ctx)
	case common.CommandSpinCW:
		var cmd common.SpinCWCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode spin_cw command: %w", err)
		}
		return driver.SpinCW(ctx)
	case common.CommandSpinCCW:
		var cmd common.SpinCCWCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode spin_ccw command: %w", err)
		}
		return driver.SpinCCW(ctx)
	case common.CommandStop:
		var cmd common.StopCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode stop command: %w", err)
		}
		return driver.Stop(ctx)
	case common.CommandThrottle:
		var cmd common.ThrottleCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode throttle command: %w", err)
		}
		_, err := driver.Throttle(ctx, cmd.Value)
		return err
	default:
		return fmt.Errorf("unsupported command type=%q", envelope.Type)
	}
}

func subscriber(ctx context.Context, driver motor.Driver) func(_ mqtt.Client, msg mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		if err := handleMotorCommand(ctx, driver, msg.Payload()); err != nil {
			log.Printf("failed to handle command topic=%s payload=%q err=%v", msg.Topic(), msg.Payload(), err)
		}
	}
}

func createDriver(name string) (motor.Driver, error) {
	switch name {
	case "dummy":
		return motor.DummyDriver{}, nil
	case "gobot":
		driver, err := motor.NewGobotDriver()
		if err != nil {
			return nil, fmt.Errorf("init gobot driver: %w", err)
		}
		return driver, nil
	case "periph":
		return motor.PeriphDriver{}, nil
	default:
		return nil, fmt.Errorf("unsupported MOTOR_DRIVER=%q", name)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// driver configuration
	driverName := common.EnvOrDefault("MOTOR_DRIVER", defaultDriver)
	driver, err := createDriver(driverName)
	if err != nil {
		log.Fatalf("failed to resolve motor driver: %v", err)
	}

	// driver cleanup
	defer func() {
		if err := driver.Close(ctx); err != nil {
			log.Printf("failed to close motor driver: %v", err)
		}
	}()

	// mqtt configuration
	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("motor-control-%d", time.Now().UnixNano()))
	topic := common.EnvOrDefault("MQTT_TOPIC", defaultTopic)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	// mqtt handlers
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("connected to broker=%s", brokerURL)
		token := client.Subscribe(topic, 1, subscriber(ctx, driver))
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("failed to subscribe topic=%s err=%v", topic, err)
			return
		}
		log.Printf("subscribed topic=%s", topic)
	})

	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("connection lost: %v", err)
	})

	// mqtt connection
	client := mqtt.NewClient(opts)
	connectToken := client.Connect()
	connectToken.Wait()
	if err := connectToken.Error(); err != nil {
		log.Fatalf("failed to connect to broker=%s err=%v", brokerURL, err)
	}

	defer client.Disconnect(250)

	<-ctx.Done()
	log.Println("shutting down motor-control")
}
