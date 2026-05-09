package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	workflowID string
	send       chan []byte
	stop       chan struct{}
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.Unlock()

		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.stop)
				close(client.send)
			}
			h.Unlock()

		case message := <-h.broadcast:
			h.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.RUnlock()
		}
	}
}

func (h *Hub) BroadcastForWorkflow(workflowID string, data []byte) {
	h.RLock()
	defer h.RUnlock()

	for client := range h.clients {
		if client.workflowID == workflowID {
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

var port = flag.String("port", "8081", "WebSocket server port")

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		workflowID: workflowID,
		send:       make(chan []byte, 256),
		stop:       make(chan struct{}),
	}

	hub.register <- client

	go func() {
		defer func() {
			hub.unregister <- client
			_ = conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			log.Printf("Received from client: %s", string(message))
		}
	}()

	go func() {
		for {
			select {
			case <-client.stop:
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			case message, ok := <-client.send:
				if !ok {
					_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("Writer error: %v", err)
					return
				}
				_, _ = w.Write(message)

				if err := w.Close(); err != nil {
					log.Printf("Close writer error: %v", err)
					return
				}
			}
		}
	}()

	log.Printf("Client connected for workflow: %s", workflowID)
}

func handleNotify(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		WorkflowID      string `json:"workflow_id"`
		Activity         string `json:"activity"`
		Status           string `json:"status"`
		TotalActivities  int    `json:"total_activities"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Received notification: %s %s %s total=%d", payload.WorkflowID, payload.Activity, payload.Status, payload.TotalActivities)

	data, _ := json.Marshal(payload)
	hub.BroadcastForWorkflow(payload.WorkflowID, data)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	flag.Parse()

	hub := NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.HandleFunc("/ws/notify", func(w http.ResponseWriter, r *http.Request) {
		handleNotify(hub, w, r)
	})

	log.Printf("WebSocket server starting on :%s", *port)
	log.Println("Endpoints:")
	log.Println("  /ws?workflow_id=X   - WebSocket connection")
	log.Println("  /ws/notify          - Receive activity notifications from workflow")
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}