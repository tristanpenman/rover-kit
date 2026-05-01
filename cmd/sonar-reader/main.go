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
	"rover-kit/pkg/sonar"

	// third-party
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL = "tcp://localhost:1883"
	defaultProvider  = "dummy"
	defaultTopic     = "rover/sonar/sample"
	defaultUartPort  = "/dev/ttyUSB0"
)

func createProvider(name string) (sonar.Provider, error) {
	switch name {
	case "dummy":
		return &sonar.DummyProvider{}, nil
	case "periph":
		provider, err := sonar.NewPeriphProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to create sonar provider: %w", err)
		}
		return provider, nil
	case "uart":
		uartPort := common.EnvOrDefault("SONAR_UART_PORT", defaultUartPort)
		provider, err := sonar.NewUartProvider(uartPort)
		if err != nil {
			return nil, fmt.Errorf("failed to create sonar provider: %w", err)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unsupported SONAR_PROVIDER=%q", name)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// provider configuration
	providerName := common.EnvOrDefault("SONAR_PROVIDER", defaultProvider)
	provider, err := createProvider(providerName)
	if err != nil {
		log.Fatalf("failed to resolve motor driver: %v", err)
	}

	// provider cleanup
	defer func() {
		if err := provider.Close(ctx); err != nil {
			log.Printf("failed to close motor driver: %v", err)
		}
	}()

	// mqtt configuration
	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("motor-control-%d", time.Now().UnixNano()))
	_ = common.EnvOrDefault("MQTT_TOPIC", defaultTopic)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	// mqtt handlers
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("connected to broker=%s", brokerURL)

		// start reading from sonar provider
		c := provider.Open(ctx)
		for reading := range c {
			log.Println("Received:", reading)
			// convert reading to JSON
			jsonReading, err := json.Marshal(reading)
			if err != nil {
				log.Printf("failed to marshal reading: %v, error: %v", reading, err)
				continue
			}
			token := client.Publish(defaultTopic, 0, false, jsonReading)
			if token.Wait() && token.Error() != nil {
				log.Printf("failed to publish reading: %v, error: %v", reading, token.Error())
			}
		}

		// shut down after provider channel is closed
		log.Println("sonar provider channel closed")
		cancel()
	})

	// shut down if connection is lost
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("connection lost: %v", err)
		cancel()
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
	log.Println("shutting down sonar-reader")
}
