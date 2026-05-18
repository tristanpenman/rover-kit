package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"rover-kit/pkg/camera"
	"rover-kit/pkg/common"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultBrokerURL     = "tcp://localhost:1883"
	defaultProvider      = "dummy"
	defaultTopic         = "rover/camera/frame"
	defaultIntervalMS    = 1000
	defaultCaptureMS     = 1000
	defaultCameraCommand = "libcamera-still"
)

func createProvider(name string) (camera.Provider, error) {
	interval, err := durationFromEnvMS("CAMERA_INTERVAL_MS", defaultIntervalMS)
	if err != nil {
		return nil, err
	}

	switch name {
	case "dummy":
		return &camera.DummyProvider{Interval: interval}, nil
	case "libcamera":
		timeout, err := durationFromEnvMS("CAMERA_CAPTURE_TIMEOUT_MS", defaultCaptureMS)
		if err != nil {
			return nil, err
		}
		width, err := intFromEnv("CAMERA_WIDTH", 0)
		if err != nil {
			return nil, err
		}
		height, err := intFromEnv("CAMERA_HEIGHT", 0)
		if err != nil {
			return nil, err
		}
		return &camera.LibcameraProvider{
			CommandPath: common.EnvOrDefault("LIBCAMERA_STILL_PATH", defaultCameraCommand),
			Interval:    interval,
			Timeout:     timeout,
			Width:       width,
			Height:      height,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported CAMERA_PROVIDER=%q", name)
	}
}

func durationFromEnvMS(key string, defaultValue int) (time.Duration, error) {
	value, err := intFromEnv(key, defaultValue)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}
	return time.Duration(value) * time.Millisecond, nil
}

func intFromEnv(key string, defaultValue int) (int, error) {
	raw := common.EnvOrDefault(key, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s=%q", key, raw)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be >= 0", key)
	}
	return value, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	providerName := common.EnvOrDefault("CAMERA_PROVIDER", defaultProvider)
	provider, err := createProvider(providerName)
	if err != nil {
		log.Fatalf("failed to resolve camera provider: %v", err)
	}

	defer func() {
		if err := provider.Close(ctx); err != nil {
			log.Printf("failed to close camera provider: %v", err)
		}
	}()

	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("camera-reader-%d", time.Now().UnixNano()))
	topic := common.EnvOrDefault("MQTT_TOPIC", defaultTopic)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("connected to broker=%s", brokerURL)

		c := provider.Open(ctx)
		for frame := range c {
			jsonFrame, err := json.Marshal(frame)
			if err != nil {
				log.Printf("failed to marshal frame captured_at=%s err=%v", frame.Timestamp.Format(time.RFC3339), err)
				continue
			}
			token := client.Publish(topic, 0, false, jsonFrame)
			if token.Wait() && token.Error() != nil {
				log.Printf("failed to publish frame captured_at=%s err=%v", frame.Timestamp.Format(time.RFC3339), token.Error())
			}
		}

		log.Println("camera provider channel closed")
		cancel()
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("connection lost: %v", err)
		cancel()
	})

	client := mqtt.NewClient(opts)
	connectToken := client.Connect()
	connectToken.Wait()
	if err := connectToken.Error(); err != nil {
		log.Fatalf("failed to connect to broker=%s err=%v", brokerURL, err)
	}

	defer client.Disconnect(250)

	<-ctx.Done()
	log.Println("shutting down camera-reader")
}
