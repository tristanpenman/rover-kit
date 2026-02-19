package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rover-kit/pkg/common"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL = "tcp://localhost:1883"
	defaultTopic     = "rover/motor/command"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	driver := common.DummyDriver{}

	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	topic := common.EnvOrDefault("MQTT_TOPIC", defaultTopic)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("motor-control-%d", time.Now().UnixNano()))

	for {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(brokerURL)
		opts.SetClientID(clientID)
		opts.SetDefaultPublishHandler(messagePubHandler)
		opts.SetOnConnectHandler(connectHandler)
		opts.SetConnectionLostHandler(connectLostHandler)

		client := mqtt.NewClient(opts)
		token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
			fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
			if string(msg.Payload()) == "spin_ccw" {
				err := driver.SpinCCW(ctx)
				if err != nil {
					return
				}
			}
		})

		if token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
}
