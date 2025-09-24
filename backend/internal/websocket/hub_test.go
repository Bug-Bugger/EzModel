package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HubTestSuite struct {
	suite.Suite
	hub *Hub
}

func (suite *HubTestSuite) SetupTest() {
	suite.hub = NewHub()
}

func (suite *HubTestSuite) TearDownTest() {
	if suite.hub != nil {
		suite.hub.Shutdown()
	}
}

func TestHubSuite(t *testing.T) {
	suite.Run(t, new(HubTestSuite))
}

// Test Hub Creation
func (suite *HubTestSuite) TestNewHub() {
	hub := NewHub()

	assert.NotNil(suite.T(), hub.projects)
	assert.NotNil(suite.T(), hub.register)
	assert.NotNil(suite.T(), hub.unregister)
	assert.NotNil(suite.T(), hub.broadcast)
	assert.NotNil(suite.T(), hub.ticker)
	assert.NotNil(suite.T(), hub.done)
}

// Test Client Registration
func (suite *HubTestSuite) TestRegisterClient() {
	// Create test client
	projectID := uuid.New()
	userID := uuid.New()
	client := suite.createTestClient(projectID, userID)

	// Start hub in goroutine
	go suite.hub.Run()
	defer suite.hub.Shutdown()

	// Register client
	suite.hub.RegisterClient(client)

	// Wait a moment for registration to process
	time.Sleep(10 * time.Millisecond)

	// Verify client count
	count := suite.hub.GetActiveClients(projectID)
	assert.Equal(suite.T(), 1, count)

	// Verify active users
	users := suite.hub.GetActiveUsers(projectID)
	assert.Len(suite.T(), users, 1)
	assert.Equal(suite.T(), userID, users[0].UserID)
}

// Test Client Unregistration
func (suite *HubTestSuite) TestUnregisterClient() {
	// Create test client
	projectID := uuid.New()
	userID := uuid.New()
	client := suite.createTestClient(projectID, userID)

	// Start hub in goroutine
	go suite.hub.Run()
	defer suite.hub.Shutdown()

	// Register client
	suite.hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Verify registration
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(projectID))

	// Unregister client
	suite.hub.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Verify unregistration
	assert.Equal(suite.T(), 0, suite.hub.GetActiveClients(projectID))
}

// Test Message Broadcasting
func (suite *HubTestSuite) TestBroadcastToProject() {
	// Create test clients for same project
	projectID := uuid.New()
	client1 := suite.createTestClient(projectID, uuid.New())
	client2 := suite.createTestClient(projectID, uuid.New())

	// Start hub in goroutine
	go suite.hub.Run()
	defer suite.hub.Shutdown()

	// Register clients
	suite.hub.RegisterClient(client1)
	suite.hub.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)

	// Create test message
	payload := UserCursorPayload{
		UserID:    client1.UserID,
		Username:  client1.Username,
		UserColor: client1.UserColor,
		CursorX:   100.5,
		CursorY:   200.5,
	}
	message, err := NewWebSocketMessage(MessageTypeUserCursor, payload, client1.UserID, projectID)
	assert.NoError(suite.T(), err)

	// Broadcast message
	suite.hub.BroadcastToProject(projectID, message, client1)

	// Check that client2 received the message (client1 should not receive its own message)
	select {
	case receivedMsg := <-client2.Send:
		var receivedMessage WebSocketMessage
		err := json.Unmarshal(receivedMsg, &receivedMessage)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), MessageTypeUserCursor, receivedMessage.Type)
	case <-time.After(100 * time.Millisecond):
		suite.T().Error("Expected to receive broadcasted message")
	}

	// Check that client1 did not receive its own message
	select {
	case <-client1.Send:
		suite.T().Error("Client should not receive its own message")
	case <-time.After(50 * time.Millisecond):
		// This is expected
	}
}

// Test Project Isolation
func (suite *HubTestSuite) TestProjectIsolation() {
	// Create clients for different projects
	project1ID := uuid.New()
	project2ID := uuid.New()
	client1 := suite.createTestClient(project1ID, uuid.New())
	client2 := suite.createTestClient(project2ID, uuid.New())

	// Start hub in goroutine
	go suite.hub.Run()
	defer suite.hub.Shutdown()

	// Register clients
	suite.hub.RegisterClient(client1)
	suite.hub.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)

	// Create test message for project1
	payload := UserCursorPayload{
		UserID:    client1.UserID,
		Username:  client1.Username,
		UserColor: client1.UserColor,
		CursorX:   100.5,
		CursorY:   200.5,
	}
	message, err := NewWebSocketMessage(MessageTypeUserCursor, payload, client1.UserID, project1ID)
	assert.NoError(suite.T(), err)

	// Broadcast to project1
	suite.hub.BroadcastToProject(project1ID, message, client1)

	// Client2 (in different project) should not receive the message
	select {
	case <-client2.Send:
		suite.T().Error("Client in different project should not receive message")
	case <-time.After(50 * time.Millisecond):
		// This is expected
	}

	// Verify project client counts
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(project1ID))
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(project2ID))
}

// Test Multiple Clients in Same Project
func (suite *HubTestSuite) TestMultipleClientsInProject() {
	projectID := uuid.New()
	client1 := suite.createTestClient(projectID, uuid.New())
	client2 := suite.createTestClient(projectID, uuid.New())
	client3 := suite.createTestClient(projectID, uuid.New())

	// Start hub in goroutine
	go suite.hub.Run()
	defer suite.hub.Shutdown()

	// Register all clients
	suite.hub.RegisterClient(client1)
	suite.hub.RegisterClient(client2)
	suite.hub.RegisterClient(client3)
	time.Sleep(10 * time.Millisecond)

	// Verify all clients are registered
	assert.Equal(suite.T(), 3, suite.hub.GetActiveClients(projectID))

	// Verify active users
	users := suite.hub.GetActiveUsers(projectID)
	assert.Len(suite.T(), users, 3)
}

// Test Hub Shutdown
func (suite *HubTestSuite) TestHubShutdown() {
	projectID := uuid.New()
	client := suite.createTestClient(projectID, uuid.New())

	// Start hub in goroutine
	go suite.hub.Run()

	// Register client
	suite.hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Verify client is registered
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(projectID))

	// Shutdown hub
	suite.hub.Shutdown()
	time.Sleep(10 * time.Millisecond)

	// Verify client count is 0 after shutdown
	assert.Equal(suite.T(), 0, suite.hub.GetActiveClients(projectID))
}

// Helper function to create a test client
func (suite *HubTestSuite) createTestClient(projectID, userID uuid.UUID) *Client {
	return &Client{
		ID:        uuid.New(),
		UserID:    userID,
		ProjectID: projectID,
		Username:  "testuser",
		UserColor: "#FF6B6B",
		Conn:      nil, // We don't need actual WebSocket connection for these tests
		Send:      make(chan []byte, 256),
		Hub:       suite.hub,
		LastPing:  time.Now(),
	}
}

// Benchmark tests for performance
func BenchmarkHubRegisterClient(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	projectID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := &Client{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ProjectID: projectID,
			Username:  "testuser",
			UserColor: "#FF6B6B",
			Send:      make(chan []byte, 256),
			Hub:       hub,
			LastPing:  time.Now(),
		}
		hub.RegisterClient(client)
	}
}

func BenchmarkHubBroadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	projectID := uuid.New()
	// Register 100 clients
	for i := 0; i < 100; i++ {
		client := &Client{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			ProjectID: projectID,
			Username:  "testuser",
			UserColor: "#FF6B6B",
			Send:      make(chan []byte, 256),
			Hub:       hub,
			LastPing:  time.Now(),
		}
		hub.RegisterClient(client)
	}

	time.Sleep(100 * time.Millisecond) // Let clients register

	payload := UserCursorPayload{
		UserID:    uuid.New(),
		Username:  "sender",
		UserColor: "#FF6B6B",
		CursorX:   100.5,
		CursorY:   200.5,
	}
	message, _ := NewWebSocketMessage(MessageTypeUserCursor, payload, uuid.New(), projectID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastToProject(projectID, message, nil)
	}
}