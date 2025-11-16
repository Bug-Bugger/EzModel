package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/config"
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

type WebSocketHandler struct {
	config         *config.Config
	hub            *websocketPkg.Hub
	jwtService     services.JWTServiceInterface
	userService    services.UserServiceInterface
	projectService services.ProjectServiceInterface
	tableService   services.TableServiceInterface
	upgrader       websocket.Upgrader
}

func NewWebSocketHandler(
	cfg *config.Config,
	hub *websocketPkg.Hub,
	jwtService services.JWTServiceInterface,
	userService services.UserServiceInterface,
	projectService services.ProjectServiceInterface,
	tableService services.TableServiceInterface,
) *WebSocketHandler {
	h := &WebSocketHandler{
		config:         cfg,
		hub:            hub,
		jwtService:     jwtService,
		userService:    userService,
		projectService: projectService,
		tableService:   tableService,
	}

	// Initialize upgrader with origin validation
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}

	return h
}

// checkOrigin validates WebSocket connection origins against allowed origins
func (h *WebSocketHandler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	if origin == "" {
		log.Printf("WebSocket: No origin header present")
		return false
	}

	// Check if origin is in allowed list
	for _, allowed := range h.config.AllowedOrigins {
		if origin == allowed {
			log.Printf("WebSocket: Accepted connection from authorized origin: %s", origin)
			return true
		}
	}

	// Log rejected origin for security monitoring
	log.Printf("WebSocket: Rejected connection from unauthorized origin: %s (allowed: %v)", origin, h.config.AllowedOrigins)
	return false
}

// HandleWebSocket upgrades HTTP connection to WebSocket for real-time collaboration
// Authentication is done via auth message after connection is established
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket: Connection attempt from %s to %s", r.RemoteAddr, r.URL.String())

	// Get project ID from URL
	projectIDStr := chi.URLParam(r, "project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		log.Printf("WebSocket: Invalid project ID '%s': %v", projectIDStr, err)
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Wait for authentication message
	log.Printf("WebSocket: Connection established, waiting for authentication message")
	h.handleUnauthenticatedConnection(conn, r, projectID)
}

// handleUnauthenticatedConnection waits for an auth message on a new WebSocket connection
func (h *WebSocketHandler) handleUnauthenticatedConnection(conn *websocket.Conn, r *http.Request, projectID uuid.UUID) {
	// Set read deadline for authentication (10 seconds)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Wait for auth message
	_, messageBytes, err := conn.ReadMessage()
	if err != nil {
		log.Printf("WebSocket: Failed to read auth message: %v", err)
		h.sendErrorAndClose(conn, "Authentication timeout or failed to read message")
		return
	}

	// Parse message
	var message websocketPkg.WebSocketMessage
	if err := json.Unmarshal(messageBytes, &message); err != nil {
		log.Printf("WebSocket: Failed to parse auth message: %v", err)
		h.sendErrorAndClose(conn, "Invalid message format")
		return
	}

	// Verify it's an auth message
	if message.Type != websocketPkg.MessageTypeAuth {
		log.Printf("WebSocket: Expected auth message, got: %s", message.Type)
		h.sendErrorAndClose(conn, "Authentication required. Send auth message first")
		return
	}

	// Parse auth payload
	var authPayload websocketPkg.AuthPayload
	if err := message.UnmarshalData(&authPayload); err != nil {
		log.Printf("WebSocket: Failed to parse auth payload: %v", err)
		h.sendErrorAndClose(conn, "Invalid auth payload")
		return
	}

	token := strings.TrimSpace(authPayload.Token)
	if token == "" {
		var fallbackErr error
		token, fallbackErr = h.extractTokenFromRequest(r)
		if fallbackErr != nil {
			log.Printf("WebSocket: No authentication token provided in message or cookies: %v", fallbackErr)
			h.sendErrorAndClose(conn, "Authentication failed: no token provided")
			return
		}
	}

	// Authenticate user with token
	user, err := h.authenticateToken(token)
	if err != nil {
		log.Printf("WebSocket: Authentication failed: %v", err)
		h.sendErrorAndClose(conn, "Authentication failed: "+err.Error())
		return
	}

	// Verify user has access to the project
	if err := h.verifyProjectAccess(user.ID, projectID); err != nil {
		log.Printf("WebSocket: Access denied: %v", err)
		h.sendErrorAndClose(conn, err.Error())
		return
	}

	log.Printf("WebSocket: Authentication successful for user %s (%s)", user.Username, user.ID)

	// Send auth success message
	h.sendAuthSuccess(conn, user.ID)

	// Register authenticated client
	h.registerAuthenticatedClient(conn, user, projectID)
}

// extractTokenFromRequest attempts to read a JWT token from cookies or headers
func (h *WebSocketHandler) extractTokenFromRequest(r *http.Request) (string, error) {
	if r == nil {
		return "", fmt.Errorf("request context unavailable")
	}

	if cookie, err := r.Cookie("access_token"); err == nil {
		if value := strings.TrimSpace(cookie.Value); value != "" {
			return value, nil
		}
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if value := strings.TrimSpace(parts[1]); value != "" {
				return value, nil
			}
		}
	}

	return "", fmt.Errorf("no authentication token found in cookies or headers")
}

// authenticateToken validates a JWT token and returns the user
func (h *WebSocketHandler) authenticateToken(token string) (*models.User, error) {
	// Validate token
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		if err == services.ErrExpiredToken {
			return nil, fmt.Errorf("token has expired")
		}
		return nil, fmt.Errorf("invalid token")
	}

	// Get user information
	user, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// verifyProjectAccess checks if user has access to the project
func (h *WebSocketHandler) verifyProjectAccess(userID, projectID uuid.UUID) error {
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("project not found")
	}

	// Check if user is owner or collaborator
	hasAccess := project.OwnerID == userID
	if !hasAccess {
		for _, collaborator := range project.Collaborators {
			if collaborator.ID == userID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return fmt.Errorf("access denied to project")
	}

	return nil
}

// sendErrorAndClose sends an error message and closes the connection
func (h *WebSocketHandler) sendErrorAndClose(conn *websocket.Conn, message string) {
	errorMsg := websocketPkg.ErrorPayload{
		Message: message,
	}
	msgBytes, err := json.Marshal(map[string]interface{}{
		"type": websocketPkg.MessageTypeError,
		"data": errorMsg,
	})
	if err != nil {
		log.Printf("WebSocket: Failed to marshal error message: %v", err)
		conn.Close()
		return
	}
	conn.WriteMessage(websocket.TextMessage, msgBytes)
	conn.Close()
}

// sendAuthSuccess sends an authentication success message
func (h *WebSocketHandler) sendAuthSuccess(conn *websocket.Conn, userID uuid.UUID) {
	successMsg := websocketPkg.AuthSuccessPayload{
		Message: "Authentication successful",
		UserID:  userID.String(),
	}
	msgBytes, _ := json.Marshal(map[string]interface{}{
		"type": websocketPkg.MessageTypeAuth,
		"data": successMsg,
	})
	conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// registerAuthenticatedClient creates and registers an authenticated client
func (h *WebSocketHandler) registerAuthenticatedClient(conn *websocket.Conn, user *models.User, projectID uuid.UUID) {
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
		LastPing:  time.Now(),
	}

	// Register client with hub
	h.hub.RegisterClient(client)

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
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
	case websocketPkg.MessageTypeTableUpdated:
		h.handleTableUpdate(client, message)
	case websocketPkg.MessageTypeTableMoved:
		h.handleTableMove(client, message)
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

	// Broadcast to all clients in the project including sender
	h.hub.BroadcastToProject(client.ProjectID, newMessage, nil)
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
		if err := h.updateProjectCanvasData(client.ProjectID, payload.CanvasData, client.UserID); err != nil {
			log.Printf("Error updating canvas data: %v", err)
		}
	}()

	// Broadcast to all clients in the project including sender
	h.hub.BroadcastToProject(client.ProjectID, message, nil)
}

// handleTableUpdate processes table position update messages
func (h *WebSocketHandler) handleTableUpdate(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	var payload websocketPkg.TablePayload
	if err := message.UnmarshalData(&payload); err != nil {
		log.Printf("Error unmarshaling table payload: %v", err)
		return
	}

	log.Printf("Table position update received: table_id=%s, position=(%f, %f)",
		payload.TableID, payload.X, payload.Y)

	// Update table position in database asynchronously
	go func() {
		if err := h.updateTablePosition(client.ProjectID, payload.TableID, payload.X, payload.Y, client.UserID); err != nil {
			log.Printf("Error updating table position: %v", err)
		}
	}()

	// Broadcast to other clients in the project (exclude sender for position updates)
	h.hub.BroadcastToProject(client.ProjectID, message, client)
}

// handleTableMove processes table position move messages (visual only, no activity)
func (h *WebSocketHandler) handleTableMove(client *websocketPkg.Client, message *websocketPkg.WebSocketMessage) {
	var payload websocketPkg.TablePayload
	if err := message.UnmarshalData(&payload); err != nil {
		log.Printf("Error unmarshaling table move payload: %v", err)
		return
	}

	log.Printf("Table position move received: table_id=%s, position=(%f, %f)",
		payload.TableID, payload.X, payload.Y)

	// Update table position in database asynchronously
	go func() {
		if err := h.updateTablePosition(client.ProjectID, payload.TableID, payload.X, payload.Y, client.UserID); err != nil {
			log.Printf("Error updating table position: %v", err)
		}
	}()

	// Broadcast visual position update to other clients only (exclude sender, no activity entries)
	h.hub.BroadcastToProject(client.ProjectID, message, client)
}

// updateProjectCanvasData updates the canvas data in the database
func (h *WebSocketHandler) updateProjectCanvasData(projectID uuid.UUID, canvasData string, userID uuid.UUID) error {
	// Use the project service to update canvas data
	_, err := h.projectService.UpdateProject(projectID, &dto.UpdateProjectRequest{
		CanvasData: &canvasData,
	}, userID)
	return err
}

// updateTablePosition updates a table's position in the database
func (h *WebSocketHandler) updateTablePosition(projectID, tableID uuid.UUID, x, y float64, userID uuid.UUID) error {
	// Use the table service to update table position
	return h.tableService.UpdateTablePosition(tableID, x, y, userID)
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
