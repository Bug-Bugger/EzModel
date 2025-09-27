package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for now
		// In production, you should validate the origin
		return true
	},
}

type WebSocketHandler struct {
	hub            *websocketPkg.Hub
	jwtService     services.JWTServiceInterface
	userService    services.UserServiceInterface
	projectService services.ProjectServiceInterface
}

func NewWebSocketHandler(
	hub *websocketPkg.Hub,
	jwtService services.JWTServiceInterface,
	userService services.UserServiceInterface,
	projectService services.ProjectServiceInterface,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		jwtService:     jwtService,
		userService:    userService,
		projectService: projectService,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket for real-time collaboration
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket: Connection attempt from %s to %s", r.RemoteAddr, r.URL.String())

	// Get project ID from URL - now using standardized "project_id" parameter
	projectIDStr := chi.URLParam(r, "project_id")
	log.Printf("WebSocket: Extracted projectIDStr from 'project_id' param: '%s'", projectIDStr)
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		log.Printf("WebSocket: UUID parsing error for '%s': %v", projectIDStr, err)
		responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Authenticate user from token
	user, err := h.authenticateWebSocketRequest(r)
	if err != nil {
		responses.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Verify user has access to the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		responses.RespondWithError(w, http.StatusNotFound, "Project not found")
		return
	}

	// Check if user is owner or collaborator
	hasAccess := project.OwnerID == user.ID
	if !hasAccess {
		for _, collaborator := range project.Collaborators {
			if collaborator.ID == user.ID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		responses.RespondWithError(w, http.StatusForbidden, "Access denied to project")
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Generate a random color for the user
	userColor := generateRandomColor()

	// Create client
	client := &websocketPkg.Client{
		ID:        uuid.New(),
		UserID:    user.ID,
		ProjectID: projectID,
		Username:  user.Username,
		UserColor: userColor,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h.hub,
	}

	// Register client with hub
	h.hub.RegisterClient(client)

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// authenticateWebSocketRequest authenticates the WebSocket connection using token from query parameter
// Browser WebSocket API doesn't support custom headers, so we use query parameters
func (h *WebSocketHandler) authenticateWebSocketRequest(r *http.Request) (*models.User, error) {
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("WebSocket: No token provided in query parameter")
		return nil, fmt.Errorf("no token provided in query parameter")
	}

	tokenLength := len(token)
	previewLength := 50
	if tokenLength < previewLength {
		previewLength = tokenLength
	}
	log.Printf("WebSocket: Received token (first 50 chars): %s...", token[:previewLength])
	log.Printf("WebSocket: Token length: %d", tokenLength)

	// Validate token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		log.Printf("WebSocket: Token validation failed: %v", err)
		if err == services.ErrExpiredToken {
			return nil, fmt.Errorf("token has expired")
		}
		return nil, fmt.Errorf("invalid token")
	}

	log.Printf("WebSocket: Token validation successful for user: %s", claims.UserID)

	// Get user information
	user, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// readPump pumps messages from the WebSocket connection to the hub
func (h *WebSocketHandler) readPump(client *websocketPkg.Client) {
	defer func() {
		h.hub.UnregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.LastPing = time.Now()
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var message websocketPkg.WebSocketMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Set message metadata
		message.UserID = client.UserID
		message.ProjectID = client.ProjectID
		message.Timestamp = time.Now()

		// Handle different message types
		h.handleMessage(client, &message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *WebSocketHandler) writePump(client *websocketPkg.Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessage(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	switch message.Type {
	case websocketPkg.MessageTypeUserCursor:
		h.handleCursorUpdate(client, message)
	case websocketPkg.MessageTypePong:
		h.handlePong(client, message)
	case websocketPkg.MessageTypeCanvasUpdated:
		h.handleCanvasUpdate(client, message)
	default:
		// For other message types, broadcast to all clients in the project
		h.hub.BroadcastToProject(client.ProjectID, message, client)
	}
}

// handleCursorUpdate processes cursor movement messages
func (h *WebSocketHandler) handleCursorUpdate(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	var payload websocketPkg.UserCursorPayload
	if err := message.UnmarshalData(&payload); err != nil {
		log.Printf("Error unmarshaling cursor payload: %v", err)
		return
	}

	// Update payload with client information
	payload.UserID = client.UserID
	payload.Username = client.Username
	payload.UserColor = client.UserColor

	// Create new message with updated payload
	newMessage, err := websocketPkg.NewWebSocketMessage(
		websocketPkg.MessageTypeUserCursor,
		payload,
		client.UserID,
		client.ProjectID,
	)
	if err != nil {
		log.Printf("Error creating cursor message: %v", err)
		return
	}

	// Broadcast to other clients in the project
	h.hub.BroadcastToProject(client.ProjectID, newMessage, client)
}

// handlePong processes pong messages for heartbeat
func (h *WebSocketHandler) handlePong(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	client.LastPing = time.Now()

	var payload websocketPkg.PongPayload
	if err := message.UnmarshalData(&payload); err != nil {
		log.Printf("Error unmarshaling pong payload: %v", err)
		return
	}

	// Update last ping time
	client.LastPing = payload.Timestamp
}

// handleCanvasUpdate processes canvas update messages
func (h *WebSocketHandler) handleCanvasUpdate(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	var payload websocketPkg.CanvasUpdatedPayload
	if err := message.UnmarshalData(&payload); err != nil {
		log.Printf("Error unmarshaling canvas payload: %v", err)
		return
	}

	// Update project canvas data in database
	// This is a simple implementation - in production you might want to
	// implement operational transformation or conflict resolution
	go func() {
		if err := h.updateProjectCanvasData(client.ProjectID, payload.CanvasData); err != nil {
			log.Printf("Error updating canvas data: %v", err)
		}
	}()

	// Broadcast to other clients in the project
	h.hub.BroadcastToProject(client.ProjectID, message, client)
}

// updateProjectCanvasData updates the canvas data in the database
func (h *WebSocketHandler) updateProjectCanvasData(projectID uuid.UUID, canvasData string) error {
	// Use the project service to update canvas data
	_, err := h.projectService.UpdateProject(projectID, &dto.UpdateProjectRequest{
		CanvasData: &canvasData,
	})
	return err
}

// generateRandomColor generates a random hex color for user identification
func generateRandomColor() string {
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FCEA2B",
		"#FF9FF3", "#54A0FF", "#5F27CD", "#00D2D3", "#FF9F43",
		"#FC427B", "#BDC581", "#82589F", "#FC9F9F", "#A3CB38",
	}
	return colors[rand.Intn(len(colors))]
}
