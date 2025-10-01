package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/redis"
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

	// Redis client for cross-region synchronization
	redisClient *redis.Client

	// Active Redis subscriptions by project ID
	subscriptions map[uuid.UUID]context.CancelFunc
	subMu         sync.Mutex
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
		projects:      make(map[uuid.UUID]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan *BroadcastMessage),
		ticker:        time.NewTicker(30 * time.Second),
		done:          make(chan struct{}),
		subscriptions: make(map[uuid.UUID]context.CancelFunc),
	}
}

// SetRedisClient sets the Redis client for cross-region synchronization
func (h *Hub) SetRedisClient(client *redis.Client) {
	h.redisClient = client
	if client != nil && client.IsEnabled() {
		log.Println("Redis client enabled for WebSocket hub - cross-region sync active")
	}
}

// Run starts the hub and handles all client connections
func (h *Hub) Run() {
	defer func() {
		h.ticker.Stop()
		h.safeCloseDoneChannel()
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
	isFirstClient := h.projects[client.ProjectID] == nil || len(h.projects[client.ProjectID]) == 0

	// Initialize project map if it doesn't exist
	if h.projects[client.ProjectID] == nil {
		h.projects[client.ProjectID] = make(map[*Client]bool)
	}

	// Add client to project
	h.projects[client.ProjectID][client] = true
	client.LastPing = time.Now()
	h.mu.Unlock()

	log.Printf("Client %s joined project %s", client.UserID, client.ProjectID)

	// Start Redis subscription if this is the first client for this project
	if isFirstClient {
		h.subscribeToRedis(client.ProjectID)
	}

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
	var shouldCloseSubscription bool

	if clients, exists := h.projects[client.ProjectID]; exists {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			// Safely close the channel
			h.safeCloseChannel(client.Send)

			// Clean up empty project maps and stop Redis subscription
			if len(clients) == 0 {
				delete(h.projects, client.ProjectID)
				shouldCloseSubscription = true
			}
		}
	}
	h.mu.Unlock()

	// Stop Redis subscription if no more clients in project
	if shouldCloseSubscription {
		h.unsubscribeFromRedis(client.ProjectID)
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

// broadcastMessage handles message broadcasting
func (h *Hub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

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
					h.safeCloseChannel(client.Send)
					delete(clients, client)
				}
			}
		}

		// Publish to Redis for cross-region synchronization (async)
		h.publishToRedis(projectID, messageBytes)
	}
}

// publishToRedis publishes a message to Redis for cross-region synchronization
func (h *Hub) publishToRedis(projectID uuid.UUID, messageBytes []byte) {
	if h.redisClient == nil || !h.redisClient.IsEnabled() {
		return
	}

	// Publish asynchronously to avoid blocking local broadcasts
	go func() {
		channel := fmt.Sprintf("project:%s", projectID.String())
		if err := h.redisClient.Publish(channel, messageBytes); err != nil {
			log.Printf("Failed to publish to Redis channel %s: %v", channel, err)
		}
	}()
}

// sendPresenceToClient sends current presence information to a specific client
func (h *Hub) sendPresenceToClient(targetClient *Client) {
	if clients, exists := h.projects[targetClient.ProjectID]; exists {
		var activeUsers []ActiveUser

		log.Printf("DEBUG: Preparing presence for client %s in project %s. Total clients in project: %d", targetClient.UserID, targetClient.ProjectID, len(clients))

		// Include ALL users in the project (including the target client)
		for client := range clients {
			log.Printf("DEBUG: Adding user %s (%s) to presence list", client.UserID, client.Username)
			activeUsers = append(activeUsers, ActiveUser{
				UserID:    client.UserID,
				Username:  client.Username,
				UserColor: client.UserColor,
				LastSeen:  client.LastPing,
			})
		}

		log.Printf("DEBUG: Sending presence with %d users to client %s", len(activeUsers), targetClient.UserID)

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
	now := time.Now()
	pingPayload := PingPayload{Timestamp: now}
	clientsToUnregister := make(map[*Client]struct{})

	h.mu.RLock()
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
				clientsToUnregister[client] = struct{}{}
				continue
			}

			select {
			case client.Send <- messageBytes:
			default:
				clientsToUnregister[client] = struct{}{}
			}
		}
	}
	h.mu.RUnlock()

	for client := range clientsToUnregister {
		h.unregisterClient(client)
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

	// Close all Redis subscriptions
	h.subMu.Lock()
	for projectID, cancel := range h.subscriptions {
		log.Printf("Closing Redis subscription for project %s", projectID)
		cancel()
	}
	h.subscriptions = make(map[uuid.UUID]context.CancelFunc)
	h.subMu.Unlock()

	// Close all client connections
	for _, clients := range h.projects {
		for client := range clients {
			h.safeCloseChannel(client.Send)
			if client.Conn != nil {
				client.Conn.Close()
			}
		}
	}

	// Clear all projects
	h.projects = make(map[uuid.UUID]map[*Client]bool)

	// Safely close the done channel
	h.safeCloseDoneChannel()
}

// subscribeToRedis subscribes to a Redis channel for cross-region messages
func (h *Hub) subscribeToRedis(projectID uuid.UUID) {
	if h.redisClient == nil || !h.redisClient.IsEnabled() {
		return
	}

	h.subMu.Lock()
	// Check if already subscribed
	if _, exists := h.subscriptions[projectID]; exists {
		h.subMu.Unlock()
		return
	}

	channel := fmt.Sprintf("project:%s", projectID.String())
	pubsub := h.redisClient.Subscribe(channel)
	if pubsub == nil {
		h.subMu.Unlock()
		return
	}

	// Create cancellable context for this subscription
	ctx, cancel := context.WithCancel(context.Background())
	h.subscriptions[projectID] = cancel
	h.subMu.Unlock()

	log.Printf("Started Redis subscription for project %s on channel %s", projectID, channel)

	// Start listening in a goroutine
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Redis subscription cancelled for project %s", projectID)
				return

			case msg, ok := <-ch:
				if !ok {
					log.Printf("Redis channel closed for project %s", projectID)
					return
				}

				// Broadcast message to local clients only (no re-publishing to Redis)
				h.broadcastFromRedis(projectID, []byte(msg.Payload))
			}
		}
	}()
}

// unsubscribeFromRedis unsubscribes from a Redis channel
func (h *Hub) unsubscribeFromRedis(projectID uuid.UUID) {
	h.subMu.Lock()
	defer h.subMu.Unlock()

	if cancel, exists := h.subscriptions[projectID]; exists {
		log.Printf("Unsubscribing from Redis for project %s", projectID)
		cancel()
		delete(h.subscriptions, projectID)
	}
}

// broadcastFromRedis broadcasts a message received from Redis to local clients
func (h *Hub) broadcastFromRedis(projectID uuid.UUID, messageBytes []byte) {
	h.mu.RLock()
	clients, exists := h.projects[projectID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	// Parse the message to check if it's from this hub
	var message WebSocketMessage
	if err := json.Unmarshal(messageBytes, &message); err != nil {
		log.Printf("Error unmarshaling Redis message: %v", err)
		return
	}

	// Broadcast to all local clients
	for client := range clients {
		select {
		case client.Send <- messageBytes:
		default:
			// Client's send channel is full, skip
			log.Printf("Skipping Redis message for client %s (channel full)", client.UserID)
		}
	}
}

// safeCloseChannel safely closes a channel if it's not already closed
func (h *Hub) safeCloseChannel(ch chan []byte) {
	defer func() {
		if recover() != nil {
			// Channel was already closed, ignore the panic
		}
	}()
	close(ch)
}

// safeCloseDoneChannel safely closes the done channel
func (h *Hub) safeCloseDoneChannel() {
	defer func() {
		if recover() != nil {
			// Channel was already closed, ignore the panic
		}
	}()
	close(h.done)
}
