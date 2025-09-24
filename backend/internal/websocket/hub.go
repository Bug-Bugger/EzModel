package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProjectID uuid.UUID
	Username  string
	UserColor string
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *Hub
	LastPing  time.Time
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by project ID
	projects map[uuid.UUID]map[*Client]bool

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from the clients
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Ticker for ping/pong heartbeat
	ticker *time.Ticker

	// Done channel for graceful shutdown
	done chan struct{}
}

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	ProjectID uuid.UUID
	Message   *WebSocketMessage
	Sender    *Client
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		projects:   make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
		ticker:     time.NewTicker(30 * time.Second),
		done:       make(chan struct{}),
	}
}

// Run starts the hub and handles all client connections
func (h *Hub) Run() {
	defer func() {
		h.ticker.Stop()
		close(h.done)
	}()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-h.ticker.C:
			h.pingClients()

		case <-h.done:
			return
		}
	}
}

// RegisterClient registers a new client to the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastToProject broadcasts a message to all clients in a project
func (h *Hub) BroadcastToProject(projectID uuid.UUID, message *WebSocketMessage, sender *Client) {
	h.broadcast <- &BroadcastMessage{
		ProjectID: projectID,
		Message:   message,
		Sender:    sender,
	}
}

// registerClient handles client registration
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Initialize project map if it doesn't exist
	if h.projects[client.ProjectID] == nil {
		h.projects[client.ProjectID] = make(map[*Client]bool)
	}

	// Add client to project
	h.projects[client.ProjectID][client] = true
	client.LastPing = time.Now()

	log.Printf("Client %s joined project %s", client.UserID, client.ProjectID)

	// Notify other clients about the new user
	userJoinedPayload := UserJoinedPayload{
		UserID:    client.UserID,
		Username:  client.Username,
		UserColor: client.UserColor,
	}

	message, err := NewWebSocketMessage(MessageTypeUserJoined, userJoinedPayload, client.UserID, client.ProjectID)
	if err != nil {
		log.Printf("Error creating user joined message: %v", err)
		return
	}

	// Broadcast to all clients in the project except the sender
	h.broadcastToProjectExcept(client.ProjectID, message, client)

	// Send current presence to the new client
	h.sendPresenceToClient(client)
}

// unregisterClient handles client disconnection
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.projects[client.ProjectID]; exists {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.Send)

			// Clean up empty project maps
			if len(clients) == 0 {
				delete(h.projects, client.ProjectID)
			}

			log.Printf("Client %s left project %s", client.UserID, client.ProjectID)

			// Notify other clients about the user leaving
			userLeftPayload := UserLeftPayload{
				UserID: client.UserID,
			}

			message, err := NewWebSocketMessage(MessageTypeUserLeft, userLeftPayload, client.UserID, client.ProjectID)
			if err != nil {
				log.Printf("Error creating user left message: %v", err)
				return
			}

			h.broadcastToProjectExcept(client.ProjectID, message, client)
		}
	}
}

// broadcastMessage handles message broadcasting
func (h *Hub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.broadcastToProjectExcept(broadcastMsg.ProjectID, broadcastMsg.Message, broadcastMsg.Sender)
}

// broadcastToProjectExcept broadcasts a message to all clients in a project except the sender
func (h *Hub) broadcastToProjectExcept(projectID uuid.UUID, message *WebSocketMessage, except *Client) {
	if clients, exists := h.projects[projectID]; exists {
		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			return
		}

		for client := range clients {
			if client != except {
				select {
				case client.Send <- messageBytes:
				default:
					// Client's send channel is full, close it
					close(client.Send)
					delete(clients, client)
				}
			}
		}
	}
}

// sendPresenceToClient sends current presence information to a specific client
func (h *Hub) sendPresenceToClient(targetClient *Client) {
	if clients, exists := h.projects[targetClient.ProjectID]; exists {
		var activeUsers []ActiveUser

		for client := range clients {
			if client != targetClient {
				activeUsers = append(activeUsers, ActiveUser{
					UserID:    client.UserID,
					Username:  client.Username,
					UserColor: client.UserColor,
					LastSeen:  client.LastPing,
				})
			}
		}

		presencePayload := UserPresencePayload{
			ActiveUsers: activeUsers,
		}

		message, err := NewWebSocketMessage(MessageTypeUserPresence, presencePayload, targetClient.UserID, targetClient.ProjectID)
		if err != nil {
			log.Printf("Error creating presence message: %v", err)
			return
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling presence message: %v", err)
			return
		}

		select {
		case targetClient.Send <- messageBytes:
		default:
			log.Printf("Failed to send presence to client %s", targetClient.UserID)
		}
	}
}

// pingClients sends ping messages to all clients for heartbeat
func (h *Hub) pingClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	now := time.Now()
	pingPayload := PingPayload{Timestamp: now}

	for projectID, clients := range h.projects {
		message, err := NewWebSocketMessage(MessageTypePing, pingPayload, uuid.Nil, projectID)
		if err != nil {
			log.Printf("Error creating ping message: %v", err)
			continue
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling ping message: %v", err)
			continue
		}

		for client := range clients {
			// Check if client is stale (no pong for 2 minutes)
			if now.Sub(client.LastPing) > 2*time.Minute {
				h.unregister <- client
				continue
			}

			select {
			case client.Send <- messageBytes:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

// GetActiveClients returns the number of active clients in a project
func (h *Hub) GetActiveClients(projectID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, exists := h.projects[projectID]; exists {
		return len(clients)
	}
	return 0
}

// GetActiveUsers returns a list of active users in a project
func (h *Hub) GetActiveUsers(projectID uuid.UUID) []ActiveUser {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var activeUsers []ActiveUser
	if clients, exists := h.projects[projectID]; exists {
		for client := range clients {
			activeUsers = append(activeUsers, ActiveUser{
				UserID:    client.UserID,
				Username:  client.Username,
				UserColor: client.UserColor,
				LastSeen:  client.LastPing,
			})
		}
	}
	return activeUsers
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all client connections
	for _, clients := range h.projects {
		for client := range clients {
			close(client.Send)
			client.Conn.Close()
		}
	}

	close(h.done)
}
