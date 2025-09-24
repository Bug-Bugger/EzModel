package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// WebSocketTestServer represents a test WebSocket server
type WebSocketTestServer struct {
	Server   *httptest.Server
	Hub      *websocketPkg.Hub
	upgrader websocket.Upgrader
}

// NewWebSocketTestServer creates a new WebSocket test server
func NewWebSocketTestServer() *WebSocketTestServer {
	hub := websocketPkg.NewHub()
	go hub.Run()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}

	server := &WebSocketTestServer{
		Hub:      hub,
		upgrader: upgrader,
	}

	// Create HTTP test server
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Simple echo server for testing
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
	}))

	return server
}

// Close shuts down the test server and hub
func (s *WebSocketTestServer) Close() {
	s.Hub.Shutdown()
	s.Server.Close()
}

// CreateTestWebSocketClient creates a test WebSocket client
func CreateTestWebSocketClient(projectID, userID uuid.UUID) *websocketPkg.Client {
	return &websocketPkg.Client{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		Username:  "testuser",
		UserColor: "#FF6B6B",
		Conn:      nil, // No actual connection needed for most tests
		Send:      make(chan []byte, 256),
		LastPing:  time.Now(),
	}
}

// CreateTestWebSocketMessage creates a test WebSocket message
func CreateTestWebSocketMessage(msgType websocketPkg.MessageType, projectID, userID uuid.UUID) *websocketPkg.WebSocketMessage {
	var payload interface{}

	switch msgType {
	case websocketPkg.MessageTypeUserCursor:
		payload = websocketPkg.UserCursorPayload{
			UserID:    userID,
			Username:  "testuser",
			UserColor: "#FF6B6B",
			CursorX:   100.5,
			CursorY:   200.5,
		}
	case websocketPkg.MessageTypeUserJoined:
		payload = websocketPkg.UserJoinedPayload{
			UserID:    userID,
			Username:  "testuser",
			UserColor: "#FF6B6B",
		}
	case websocketPkg.MessageTypeUserLeft:
		payload = websocketPkg.UserLeftPayload{
			UserID: userID,
		}
	case websocketPkg.MessageTypeTableCreated:
		payload = websocketPkg.TablePayload{
			TableID: uuid.New(),
			Name:    "Test Table",
			X:       100.0,
			Y:       200.0,
		}
	case websocketPkg.MessageTypeCanvasUpdated:
		payload = websocketPkg.CanvasUpdatedPayload{
			CanvasData: `{"tables": [], "relationships": []}`,
		}
	default:
		payload = map[string]interface{}{"test": "data"}
	}

	message, _ := websocketPkg.NewWebSocketMessage(msgType, payload, userID, projectID)
	return message
}

// ConnectWebSocket creates a WebSocket connection to a test server
func ConnectWebSocket(t *testing.T, serverURL string) *websocket.Conn {
	// Convert http:// to ws://
	wsURL := strings.Replace(serverURL, "http://", "ws://", 1)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	return conn
}

// ConnectWebSocketWithAuth creates a WebSocket connection with authentication via Authorization header
func ConnectWebSocketWithAuth(t *testing.T, serverURL, token string) *websocket.Conn {
	// Convert http:// to ws://
	wsURL := strings.Replace(serverURL, "http://", "ws://", 1)

	// Create headers with Authorization Bearer token
	headers := http.Header{}
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	require.NoError(t, err)

	return conn
}

// SendWebSocketMessage sends a message through WebSocket connection
func SendWebSocketMessage(t *testing.T, conn *websocket.Conn, message *websocketPkg.WebSocketMessage) {
	err := conn.WriteJSON(message)
	require.NoError(t, err)
}

// ReadWebSocketMessage reads a message from WebSocket connection
func ReadWebSocketMessage(t *testing.T, conn *websocket.Conn) *websocketPkg.WebSocketMessage {
	var message websocketPkg.WebSocketMessage
	err := conn.ReadJSON(&message)
	require.NoError(t, err)
	return &message
}

// ReadWebSocketMessageWithTimeout reads a message with timeout
func ReadWebSocketMessageWithTimeout(t *testing.T, conn *websocket.Conn, timeout time.Duration) *websocketPkg.WebSocketMessage {
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{}) // Clear deadline

	var message websocketPkg.WebSocketMessage
	err := conn.ReadJSON(&message)
	require.NoError(t, err)
	return &message
}

// AssertWebSocketMessage asserts properties of a WebSocket message
func AssertWebSocketMessage(t *testing.T, message *websocketPkg.WebSocketMessage, expectedType websocketPkg.MessageType, expectedUserID, expectedProjectID uuid.UUID) {
	require.Equal(t, expectedType, message.Type)
	require.Equal(t, expectedUserID, message.UserID)
	require.Equal(t, expectedProjectID, message.ProjectID)
	require.NotEmpty(t, message.Data)
	require.WithinDuration(t, time.Now(), message.Timestamp, time.Minute)
}

// WaitForChannelMessage waits for a message on a channel with timeout
func WaitForChannelMessage(t *testing.T, ch <-chan []byte, timeout time.Duration) []byte {
	select {
	case msg := <-ch:
		return msg
	case <-time.After(timeout):
		t.Fatal("Timeout waiting for channel message")
		return nil
	}
}

// CreateTestJWTToken creates a test JWT token for WebSocket authentication
func CreateTestJWTToken(t *testing.T, jwtService services.JWTServiceInterface, user *User) string {
	modelUser := &models.User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}

	tokenPair, err := jwtService.GenerateTokenPair(modelUser)
	require.NoError(t, err)
	return tokenPair.AccessToken
}

// User represents a simplified user for testing
type User struct {
	ID       uuid.UUID
	Email    string
	Username string
}

// CreateTestUserForWebSocket creates a simplified user for WebSocket tests
func CreateTestUserForWebSocket() *User {
	return &User{
		ID:       uuid.New(),
		Email:    "wstest@example.com",
		Username: "wstestuser",
	}
}

// MockWebSocketHandler creates a mock WebSocket handler for testing
func MockWebSocketHandler(hub *websocketPkg.Hub) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Extract project ID from URL (simplified for testing)
		projectID := uuid.New()
		userID := uuid.New()

		client := &websocketPkg.Client{
			ID:        uuid.New(),
			UserID:    userID,
			ProjectID: projectID,
			Username:  "testuser",
			UserColor: "#FF6B6B",
			Conn:      conn,
			Send:      make(chan []byte, 256),
			Hub:       hub,
			LastPing:  time.Now(),
		}

		hub.RegisterClient(client)
		defer hub.UnregisterClient(client)

		// Simple message echo loop
		for {
			var message websocketPkg.WebSocketMessage
			err := conn.ReadJSON(&message)
			if err != nil {
				break
			}

			// Echo message back
			err = conn.WriteJSON(&message)
			if err != nil {
				break
			}
		}
	}
}

// CreateWebSocketRequestWithAuth creates an HTTP request with WebSocket headers and auth
func CreateWebSocketRequestWithAuth(t *testing.T, url, token string, projectID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "test-key")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Add project ID to context or URL params
	ctx := context.WithValue(req.Context(), "projectID", projectID.String())
	return req.WithContext(ctx)
}
