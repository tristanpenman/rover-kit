package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	// internal
	"rover-kit/pkg/common"

	// third-party
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

const (
	defaultBrokerURL     = "tcp://localhost:1883"
	defaultMotorCmdTopic = "rover/motor/cmd"
	defaultSonarTopic    = "rover/sonar/sample"
)

type wsServer struct {
	clients       map[*websocket.Conn]struct{}
	motorCmdTopic string
	mqttClient    mqtt.Client
	mu            sync.Mutex
	upgrader      websocket.Upgrader
}

type commandEnvelope struct {
	Type common.CommandType `json:"type"`
}

type throttleResponse struct {
	Type   common.CommandType `json:"type"`
	Active bool               `json:"active"`
}

func newWSServer(motorCmdTopic string) *wsServer {
	return &wsServer{
		clients:       make(map[*websocket.Conn]struct{}),
		motorCmdTopic: motorCmdTopic,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (s *wsServer) addClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = struct{}{}
}

func (s *wsServer) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, conn)
}

func (s *wsServer) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	// websocket cleanup
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("close websocket error: %v", err)
		}
	}(conn)

	s.addClient(conn)
	defer s.removeClient(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return
		}

		_, err = parseCommand(message)
		if err != nil {
			if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); writeErr != nil {
				log.Printf("websocket write error: %v", writeErr)
				return
			}
			continue
		}

		log.Printf("forwarding message: %v", message)
		s.mqttClient.Publish(defaultMotorCmdTopic, 0, false, message)

		if writeErr := conn.WriteMessage(websocket.TextMessage, message); writeErr != nil {
			log.Printf("websocket write json error: %v", writeErr)
			return
		}
	}
}

func parseCommand(message []byte) (any, error) {
	var envelope commandEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}

	switch envelope.Type {
	case common.CommandForwards:
		return common.ForwardsCommand{Type: common.CommandForwards}, nil
	case common.CommandBackwards:
		return common.BackwardsCommand{Type: common.CommandBackwards}, nil
	case common.CommandSpinCW:
		return common.SpinCWCommand{Type: common.CommandSpinCW}, nil
	case common.CommandSpinCCW:
		return common.SpinCCWCommand{Type: common.CommandSpinCCW}, nil
	case common.CommandStop:
		return common.StopCommand{Type: common.CommandStop}, nil
	case common.CommandThrottle:
		var command common.ThrottleCommand
		if err := json.Unmarshal(message, &command); err != nil {
			return nil, fmt.Errorf("bad request: %w", err)
		}
		return throttleResponse{Type: common.CommandThrottle, Active: command.Value != 0}, nil
	default:
		return nil, errors.New("invalid payload type")
	}
}

func staticDirPath(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}

	execPath, err := os.Executable()
	if err != nil {
		return dir
	}

	return filepath.Join(filepath.Dir(execPath), dir)
}

func main() {
	host := flag.String("host", "0.0.0.0", "interface to bind")
	port := flag.Int("port", 7200, "port to bind")
	staticDir := flag.String("static-dir", "static", "path to static web assets")
	flag.Parse()

	// mqtt configuration
	brokerURL := common.EnvOrDefault("MQTT_BROKER", defaultBrokerURL)
	clientID := common.EnvOrDefault("MQTT_CLIENT_ID", fmt.Sprintf("motor-control-%d", time.Now().UnixNano()))
	motorCmdTopic := common.EnvOrDefault("MQTT_MOTOR_CMD_TOPIC", defaultMotorCmdTopic)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	server := newWSServer(motorCmdTopic)

	// mqtt handlers
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("connected to broker=%s", brokerURL)

		token := client.Subscribe(defaultSonarTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
			log.Printf("received sonar message: %s", string(msg.Payload()))

			for conn := range server.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg.Payload()); err != nil {
					log.Printf("websocket write error: %v", err)
				}
			}
		})

		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("failed to subscribe topic=%s err=%v", defaultSonarTopic, err)
			return
		}
		log.Printf("subscribed topic=%s", defaultSonarTopic)

	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("connection lost: %v", err)
	})

	// mqtt connection
	server.mqttClient = mqtt.NewClient(opts)
	connectToken := server.mqttClient.Connect()
	connectToken.Wait()
	if err := connectToken.Error(); err != nil {
		log.Fatalf("failed to connect to broker=%s err=%v", brokerURL, err)
	}

	defer server.mqttClient.Disconnect(250)

	// setup web server
	webRoot := staticDirPath(*staticDir)
	mux := http.NewServeMux()
	mux.Handle("/ws", http.HandlerFunc(server.websocketHandler))
	mux.Handle("/", http.FileServer(http.Dir(webRoot)))

	// start listening
	address := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("starting web bridge on http://%s using static dir %s", address, webRoot)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
