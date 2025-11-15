package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/config"
	mockService "github.com/Bug-Bugger/ezmodel/internal/mocks/service"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/testutil"
	websocketPkg "github.com/Bug-Bugger/ezmodel/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WebSocketHandlerTestSuite struct {
	suite.Suite
	cfg              *config.Config
	handler          *WebSocketHandler
	hub              *websocketPkg.Hub
	mockJWTService   *mockService.MockJWTService
	mockUserService  *mockService.MockUserService
	mockProjService  *mockService.MockProjectService
	mockTableService *mockService.MockTableService
	upgrader         websocket.Upgrader
}

func (suite *WebSocketHandlerTestSuite) SetupTest() {
	suite.hub = websocketPkg.NewHub()
	go suite.hub.Run()

	suite.cfg = &config.Config{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:4173",
		},
	}

	suite.mockJWTService = new(mockService.MockJWTService)
	suite.mockUserService = new(mockService.MockUserService)
	suite.mockProjService = new(mockService.MockProjectService)
	suite.mockTableService = new(mockService.MockTableService)

	suite.handler = NewWebSocketHandler(
		suite.cfg,
		suite.hub,
		suite.mockJWTService,
		suite.mockUserService,
		suite.mockProjService,
		suite.mockTableService,
	)

	suite.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func (suite *WebSocketHandlerTestSuite) TearDownTest() {
	if suite.hub != nil {
		suite.hub.Shutdown()
	}
}

func TestWebSocketHandlerSuite(t *testing.T) {
	suite.Run(t, new(WebSocketHandlerTestSuite))
}

// Test invalid project ID format
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_InvalidProjectID() {
	req := httptest.NewRequest(http.MethodGet, "/projects/invalid-uuid/collaborate", nil)
	req = suite.addURLParam(req, "project_id", "invalid-uuid")

	w := httptest.NewRecorder()

	suite.handler.HandleWebSocket(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Test authentication with invalid token
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_InvalidToken() {
	projectID := uuid.New()

	// Setup mocks
	suite.mockJWTService.On("ValidateToken", "invalid-token").
		Return(nil, services.ErrInvalidToken)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:] // Replace http with ws
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message with invalid token
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": "invalid-token",
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message received
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "Authentication failed")

	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test authentication with expired token
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_ExpiredToken() {
	projectID := uuid.New()

	// Setup mocks
	suite.mockJWTService.On("ValidateToken", "expired-token").
		Return(nil, services.ErrExpiredToken)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message with expired token
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": "expired-token",
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message received
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "token has expired")

	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test authentication with non-auth message sent first
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_NonAuthMessageFirst() {
	projectID := uuid.New()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send non-auth message (should fail)
	cursorMsg := map[string]interface{}{
		"type": "cursor_move",
		"data": map[string]interface{}{
			"x": 100,
			"y": 200,
		},
	}
	err = ws.WriteJSON(cursorMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message received
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "Authentication required")
}

// Test user not found
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_UserNotFound() {
	projectID := uuid.New()
	userID := uuid.New()
	token := "valid-token"

	// Setup mocks
	claims := &services.CustomClaims{
		UserID: userID,
		Email:  "test@example.com",
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(nil, services.ErrUserNotFound)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": token,
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "user not found")

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

// Test project not found
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_ProjectNotFound() {
	projectID := uuid.New()
	userID := uuid.New()
	token := "valid-token"

	// Setup test data
	user := testutil.CreateTestUser()
	user.ID = userID

	// Setup mocks
	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(nil, services.ErrProjectNotFound)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": token,
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "project not found")

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test user access denied (not owner or collaborator)
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_AccessDenied() {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	token := "valid-token"

	// Setup test data
	user := testutil.CreateTestUser()
	user.ID = userID

	project := testutil.CreateTestProject(ownerID)
	project.ID = projectID
	project.Collaborators = []models.User{} // User is not a collaborator

	// Setup mocks
	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(project, nil)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": token,
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read error response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert error message
	assert.Equal(suite.T(), "error", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data["message"], "access denied")

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test authentication succeeds when token is provided via Authorization header
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_TokenFromAuthorizationHeader() {
	projectID := uuid.New()
	userID := uuid.New()
	token := "header-token"

	user := testutil.CreateTestUser()
	user.ID = userID

	project := testutil.CreateTestProject(userID)
	project.ID = projectID

	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(project, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, headers)
	suite.Require().NoError(err)
	defer ws.Close()

	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{}, // No token in payload
	}
	suite.Require().NoError(ws.WriteJSON(authMsg))

	var response map[string]interface{}
	suite.Require().NoError(ws.ReadJSON(&response))
	suite.Equal("auth", response["type"])
	data := response["data"].(map[string]interface{})
	suite.Equal("Authentication successful", data["message"])
	suite.Equal(userID.String(), data["user_id"])

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test authentication succeeds when token is provided via cookie instead of message payload
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_TokenFromCookie() {
	projectID := uuid.New()
	userID := uuid.New()
	token := "cookie-token"

	user := testutil.CreateTestUser()
	user.ID = userID

	project := testutil.CreateTestProject(userID)
	project.ID = projectID

	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(project, nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	headers := http.Header{}
	headers.Set("Cookie", (&http.Cookie{Name: "access_token", Value: token}).String())

	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, headers)
	suite.Require().NoError(err)
	defer ws.Close()

	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{}, // No token in payload
	}
	suite.Require().NoError(ws.WriteJSON(authMsg))

	var response map[string]interface{}
	suite.Require().NoError(ws.ReadJSON(&response))
	suite.Equal("auth", response["type"])
	data := response["data"].(map[string]interface{})
	suite.Equal("Authentication successful", data["message"])
	suite.Equal(userID.String(), data["user_id"])

	// Ensure mocks were called
	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test successful authentication as collaborator
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_SuccessAsCollaborator() {
	projectID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	token := "valid-token"

	// Setup test data
	user := testutil.CreateTestUser()
	user.ID = userID

	project := testutil.CreateTestProject(ownerID)
	project.ID = projectID
	project.Collaborators = []models.User{*user} // User is a collaborator

	// Setup mocks
	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(project, nil)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = suite.addURLParam(r, "project_id", projectID.String())
		suite.handler.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + server.URL[4:]
	ws, err := suite.dialWebSocket(wsURL, nil)
	suite.Require().NoError(err)
	defer ws.Close()

	// Send auth message
	authMsg := map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"token": token,
		},
	}
	err = ws.WriteJSON(authMsg)
	assert.NoError(suite.T(), err)

	// Read auth success response
	var response map[string]interface{}
	err = ws.ReadJSON(&response)
	assert.NoError(suite.T(), err)

	// Assert auth success message received
	assert.Equal(suite.T(), "auth", response["type"])
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Authentication successful", data["message"])

	// Wait a moment for client registration
	time.Sleep(50 * time.Millisecond)

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Helper method to dial WebSocket connections with default allowed origin
func (suite *WebSocketHandlerTestSuite) dialWebSocket(wsURL string, headers http.Header) (*websocket.Conn, error) {
	if headers == nil {
		headers = http.Header{}
	}
	if headers.Get("Origin") == "" && len(suite.cfg.AllowedOrigins) > 0 {
		headers.Set("Origin", suite.cfg.AllowedOrigins[0])
	}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	return conn, err
}

// Helper method to add URL parameters to request
func (suite *WebSocketHandlerTestSuite) addURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}
