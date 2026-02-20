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
	driver "rover-kit/pkg/motor"

	// external
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL = "tcp://localhost:1883"
	defaultTopic     = "rover/motor/command"
)

type commandEnvelope struct {
	Type common.CommandType `json:"type"`
}

func handleMotorCommand(ctx context.Context, motorDriver driver.MotorDriver, payload []byte) error {
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
		return motorDriver.Forwards(ctx)
	case common.CommandBackwards:
		var cmd common.BackwardsCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode backwards command: %w", err)
		}
		return motorDriver.Backwards(ctx)
	case common.CommandSpinCW:
		var cmd common.SpinCWCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode spin_cw command: %w", err)
		}
		return motorDriver.SpinCW(ctx)
	case common.CommandSpinCCW:
		var cmd common.SpinCCWCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode spin_ccw command: %w", err)
		}
		return motorDriver.SpinCCW(ctx)
	case common.CommandStop:
		var cmd common.StopCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode stop command: %w", err)
		}
		return motorDriver.Stop(ctx)
	case common.CommandThrottle:
		var cmd common.ThrottleCommand
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return fmt.Errorf("decode throttle command: %w", err)
		}
		_, err := motorDriver.Throttle(ctx, cmd.Value)
		return err
	default:
		return fmt.Errorf("unsupported command type=%q", envelope.Type)
	}
}

func subscriber(ctx context.Context, driver common.DummyDriver) func(_ mqtt.Client, msg mqtt.Message) {
	return func(_ mqtt.Client, msg mqtt.Message) {
		if err := handleMotorCommand(ctx, driver, msg.Payload()); err != nil {
			log.Printf("failed to handle command topic=%s payload=%q err=%v", msg.Topic(), msg.Payload(), err)
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dummyDriver := common.DummyDriver{}
	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	topic := common.EnvOrDefault("MQTT_TOPIC", defaultTopic)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("motor-control-%d", time.Now().UnixNano()))

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("connected to broker=%s", brokerURL)
		token := client.Subscribe(topic, 1, subscriber(ctx, dummyDriver))
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
