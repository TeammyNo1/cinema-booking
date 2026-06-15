package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"cinema-booking/internal/auth"
	"cinema-booking/internal/models"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn       *gws.Conn
	showtimeID string
	userID     string
	send       chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]bool // showtimeID → set of clients
	authHandler *auth.Handler
}

func NewHub(authHandler *auth.Handler) *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Client]bool),
		authHandler: authHandler,
	}
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[client.showtimeID] == nil {
		h.rooms[client.showtimeID] = make(map[*Client]bool)
	}
	h.rooms[client.showtimeID][client] = true
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[client.showtimeID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, client.showtimeID)
		}
	}
}

// BroadcastToShowtime sends a WSMessage to all clients watching a showtime.
func (h *Hub) BroadcastToShowtime(showtimeID string, msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}
	h.mu.RLock()
	clients := h.rooms[showtimeID]
	h.mu.RUnlock()
	for client := range clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			h.unregister(client)
		}
	}
}

// HandleWS upgrades an HTTP connection to WebSocket.
func (h *Hub) HandleWS(c *gin.Context) {
	showtimeID := c.Param("showtimeId")
	if showtimeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing showtime_id"})
		return
	}

	// Optional auth from query param
	userID := "anonymous"
	if token := c.Query("token"); token != "" {
		if claims, err := h.authHandler.ValidateJWT(token); err == nil {
			userID = claims.UserID
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:       conn,
		showtimeID: showtimeID,
		userID:     userID,
		send:       make(chan []byte, 256),
	}
	h.register(client)

	// Write pump
	go func() {
		defer func() {
			conn.Close()
			h.unregister(client)
		}()
		for msg := range client.send {
			if err := conn.WriteMessage(gws.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Read pump (keep connection alive, handle pings)
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
}
