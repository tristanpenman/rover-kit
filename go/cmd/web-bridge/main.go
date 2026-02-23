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

	"github.com/gorilla/websocket"

	"rover-kit/pkg/common"
)

type wsServer struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]struct{}
	mu       sync.Mutex
}

type commandEnvelope struct {
	Type common.CommandType `json:"type"`
}

type throttleResponse struct {
	Type   common.CommandType `json:"type"`
	Active bool               `json:"active"`
}

func newWSServer() *wsServer {
	return &wsServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]struct{}),
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
	defer conn.Close()
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

		response, err := parseCommand(message)
		if err != nil {
			if writeErr := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); writeErr != nil {
				log.Printf("websocket write error: %v", writeErr)
				return
			}
			continue
		}

		if writeErr := conn.WriteJSON(response); writeErr != nil {
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
	port := flag.Int("port", 8000, "port to bind")
	staticDir := flag.String("static-dir", "web", "path to static web assets")
	flag.Parse()

	server := newWSServer()
	mux := http.NewServeMux()

	webRoot := staticDirPath(*staticDir)
	mux.Handle("/ws", http.HandlerFunc(server.websocketHandler))
	mux.Handle("/", http.FileServer(http.Dir(webRoot)))

	address := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("starting web bridge on http://%s using static dir %s", address, webRoot)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
