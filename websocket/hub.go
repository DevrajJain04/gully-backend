package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client wraps a single WebSocket connection.
type Client struct {
	conn    *websocket.Conn
	groupID string
	hub     *Hub
}

// Hub manages per-group WebSocket client sets.
type Hub struct {
	mu     sync.RWMutex
	groups map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		groups: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.groups[client.groupID] == nil {
		h.groups[client.groupID] = make(map[*Client]bool)
	}
	h.groups[client.groupID][client] = true
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.groups[client.groupID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.groups, client.groupID)
		}
	}
}

// BroadcastToGroup sends a JSON message to every client in the given group.
func (h *Hub) BroadcastToGroup(groupID string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.groups[groupID]
	h.mu.RUnlock()

	for client := range clients {
		if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("write error: %v", err)
			client.conn.Close()
			h.unregister(client)
		}
	}
}

// HandleWebSocket is the Gin handler for WebSocket upgrade at /ws/group/:groupId.
func (h *Hub) HandleWebSocket(c *gin.Context) {
	groupID := c.Param("groupId")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	client := &Client{conn: conn, groupID: groupID, hub: h}
	h.register(client)

	// Keep the connection alive; read messages (we only need pong/close frames).
	go func() {
		defer func() {
			h.unregister(client)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
