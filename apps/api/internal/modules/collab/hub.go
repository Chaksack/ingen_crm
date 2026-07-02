package collab

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

// Hub tracks live WebSocket connections per user for 1:1 message delivery.
// Phase 2 swaps this for a shared Redis-backed presence/gateway service
// when the collab module runs on more than one process.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[string]*websocket.Conn)}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = conn
}

func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) Send(userID string, v interface{}) bool {
	h.mu.RLock()
	conn, ok := h.conns[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	if err := conn.WriteJSON(v); err != nil {
		return false
	}
	return true
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.conns[userID]
	return ok
}
