package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
)

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	rooms          map[int]*Room // userID -> Room
	broadcast      chan *Message
	register       chan *Client
	unregister     chan *Client
	mu             sync.RWMutex
	pomodoroStates map[int]*Message // userID -> latest pomodoro sync message
	logger         *logger.ZapLogger
}

// NewHub creates a new WebSocket hub
func NewHub(logger *logger.ZapLogger) *Hub {
	return &Hub{
		rooms:          make(map[int]*Room),
		broadcast:      make(chan *Message, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		pomodoroStates: make(map[int]*Message),
		logger:         logger,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case message := <-h.broadcast:
			h.sendToUser(message.UserID, message)
		}
	}
}

// addClient adds a client to the appropriate room
func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.userID]
	if !exists {
		room = NewRoom(client.userID)
		h.rooms[client.userID] = room
	}
	room.AddClient(client)

	/*
		h.logger.Info("WebSocket client connected", map[string]interface{}{
			"user_id":       client.userID,
			"total_clients": room.ClientCount(),
			"action":        "WS_CLIENT_CONNECTED",
		})
	*/

	// Check for active pomodoro session
	if msg, exists := h.pomodoroStates[client.userID]; exists {
		if msg.Type == "pomodoro.sync" {
			// Check if the session has already expired
			session, ok := msg.Payload["session"].(map[string]interface{})
			if ok {
				var startTime float64
				if s, ok := session["startTime"].(float64); ok {
					startTime = s
				}

				var duration float64
				if d, ok := session["duration"].(float64); ok {
					duration = d
				}

				if startTime > 0 && duration > 0 {
					// startTime is in milliseconds
					expiry := time.Unix(int64(startTime/1000), 0).Add(time.Duration(duration) * time.Minute)
					if time.Now().After(expiry) {
						// Session expired! Send it as complete
						msg.Payload["state"] = "complete"
						// Update global state too so other late joiners get the same info
						h.pomodoroStates[client.userID] = msg
					}
				}
			}
		}
		h.SendToClient(client, msg)
	}
}

// removeClient removes a client from their room
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.rooms[client.userID]; exists {
		room.RemoveClient(client)
		close(client.send)

		if room.IsEmpty() {
			delete(h.rooms, client.userID)
		}
	}

	/*
		h.logger.Info("WebSocket client disconnected", map[string]interface{}{
			"user_id": client.userID,
			"action":  "WS_CLIENT_DISCONNECTED",
		})
	*/
}

// sendToUser sends a message to all clients of a specific user
func (h *Hub) sendToUser(userID int, message *Message) {
	h.mu.RLock()
	room, exists := h.rooms[userID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal WebSocket message", err, map[string]interface{}{
			"user_id": userID,
			"type":    message.Type,
			"action":  "WS_MESSAGE_MARSHAL_FAILED",
		})
		return
	}

	room.Broadcast(data, message.ExcludeCID, h.logger)

	/*
		h.logger.Info("WebSocket message sent", map[string]interface{}{
			"user_id": userID,
			"type":    message.Type,
			"action":  "WS_MESSAGE_SENT",
		})
	*/
}

// PublishToUser queues a message for broadcast to a specific user (non-blocking)
// If excludeCID is provided, that client will be skipped
func (h *Hub) PublishToUser(userID int, msg *Message) {
	msg.UserID = userID

	// Manage Pomodoro state persistence
	if msg.Type == "pomodoro.sync" {
		h.mu.Lock()
		state, ok := msg.Payload["state"].(string)
		if ok && state == "idle" {
			delete(h.pomodoroStates, userID)
		} else {
			h.pomodoroStates[userID] = msg
		}
		h.mu.Unlock()
	}

	select {
	case h.broadcast <- msg:
		// Message queued
	default:
		h.logger.Error("Broadcast channel full, dropping message", nil, map[string]interface{}{
			"user_id": userID,
			"type":    msg.Type,
			"action":  "WS_BROADCAST_CHANNEL_FULL",
		})
	}
}

// BroadcastToUser sends a message directly to a user (blocking)
func (h *Hub) BroadcastToUser(userID int, message *Message) {
	message.UserID = userID
	h.sendToUser(userID, message)
}

// BroadcastToUsers sends a message to multiple users
func (h *Hub) BroadcastToUsers(userIDs []int, message *Message) {
	for _, userID := range userIDs {
		h.BroadcastToUser(userID, message)
	}
}

// BroadcastToAll sends a message to all connected users
func (h *Hub) BroadcastToAll(message *Message) {
	h.mu.RLock()
	userIDs := make([]int, 0, len(h.rooms))
	for userID := range h.rooms {
		userIDs = append(userIDs, userID)
	}
	h.mu.RUnlock()

	for _, userID := range userIDs {
		h.BroadcastToUser(userID, message)
	}
}

// GetActiveConnections returns the number of active connections for a user
func (h *Hub) GetActiveConnections(userID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, exists := h.rooms[userID]; exists {
		return room.ClientCount()
	}
	return 0
}

// GetTotalConnections returns the total number of active connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, room := range h.rooms {
		total += room.ClientCount()
	}
	return total
}

// GetConnectedUserCount returns the number of users with active connections
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// Register adds a client to the hub (called externally)
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub (called externally)
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(client *Client, message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal WebSocket message for client", err, map[string]interface{}{
			"user_id": client.userID,
			"type":    message.Type,
			"action":  "WS_MESSAGE_MARSHAL_FAILED",
		})
		return
	}

	select {
	case client.send <- data:
		// Message queued successfully
	default:
		h.logger.Error("Client send buffer full", nil, map[string]interface{}{
			"user_id": client.userID,
			"action":  "WS_BUFFER_FULL",
		})
	}
}
