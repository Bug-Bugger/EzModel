package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/models"
	mockService "github.com/Bug-Bugger/ezmodel/internal/mocks/service"
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
	handler         *WebSocketHandler
	hub             *websocketPkg.Hub
	mockJWTService  *mockService.MockJWTService
	mockUserService *mockService.MockUserService
	mockProjService *mockService.MockProjectService
	upgrader        websocket.Upgrader
}

func (suite *WebSocketHandlerTestSuite) SetupTest() {
	suite.hub = websocketPkg.NewHub()
	go suite.hub.Run()

	suite.mockJWTService = new(mockService.MockJWTService)
	suite.mockUserService = new(mockService.MockUserService)
	suite.mockProjService = new(mockService.MockProjectService)

	suite.handler = NewWebSocketHandler(
		suite.hub,
		suite.mockJWTService,
		suite.mockUserService,
		suite.mockProjService,
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

// Test authentication with invalid token
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_InvalidToken() {
	projectID := uuid.New()

	// Setup mocks
	suite.mockJWTService.On("ValidateToken", "invalid-token").
		Return(nil, services.ErrInvalidToken)

	// Create request
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", "invalid-token", projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test authentication with expired token
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_ExpiredToken() {
	projectID := uuid.New()

	// Setup mocks
	suite.mockJWTService.On("ValidateToken", "expired-token").
		Return(nil, services.ErrExpiredToken)

	// Create request
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", "expired-token", projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test authentication with missing token
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_MissingToken() {
	projectID := uuid.New()

	// Create request without token
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", "", projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
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

	// Create request
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", token, projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
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

	// Create request
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", token, projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
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

	// Create request
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/"+projectID.String()+"/collaborate", token, projectID)
	req = suite.addURLParam(req, "projectID", projectID.String())

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test successful authentication as project owner
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_SuccessAsOwner() {
	projectID := uuid.New()
	userID := uuid.New()
	token := "valid-token"

	// Setup test data
	user := testutil.CreateTestUser()
	user.ID = userID

	project := testutil.CreateTestProject(userID) // User is the owner
	project.ID = projectID

	// Setup mocks
	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}
	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)
	suite.mockUserService.On("GetUserByID", userID).Return(user, nil)
	suite.mockProjService.On("GetProjectByID", projectID).Return(project, nil)

	// Create test server with proper routing
	router := chi.NewRouter()
	router.Get("/projects/{projectID}/collaborate", suite.handler.HandleWebSocket)
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	conn := testutil.ConnectWebSocketWithAuth(suite.T(), server.URL+"/projects/"+projectID.String()+"/collaborate", token)
	defer conn.Close()

	// Wait a moment for connection to establish
	time.Sleep(50 * time.Millisecond)

	// Verify client is registered in hub
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(projectID))

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

	// Create test server with proper routing
	router := chi.NewRouter()
	router.Get("/projects/{projectID}/collaborate", suite.handler.HandleWebSocket)
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket
	conn := testutil.ConnectWebSocketWithAuth(suite.T(), server.URL+"/projects/"+projectID.String()+"/collaborate", token)
	defer conn.Close()

	// Wait a moment for connection to establish
	time.Sleep(50 * time.Millisecond)

	// Verify client is registered in hub
	assert.Equal(suite.T(), 1, suite.hub.GetActiveClients(projectID))

	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockProjService.AssertExpectations(suite.T())
}

// Test invalid project ID format
func (suite *WebSocketHandlerTestSuite) TestHandleWebSocket_InvalidProjectID() {
	token := "valid-token"

	// Create request with invalid project ID
	req := testutil.CreateWebSocketRequestWithAuth(suite.T(), "/projects/invalid-uuid/collaborate", token, uuid.New())
	req = suite.addURLParam(req, "projectID", "invalid-uuid")

	w := httptest.NewRecorder()

	// Execute
	suite.handler.HandleWebSocket(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Test authentication with invalid authorization header format
func (suite *WebSocketHandlerTestSuite) TestAuthenticateWebSocketRequest_InvalidFormat() {
	// Create request with invalid authorization header format
	req := httptest.NewRequest(http.MethodGet, "/collaborate", nil)
	req.Header.Set("Authorization", "InvalidFormat token")

	// Execute
	result, err := suite.handler.authenticateWebSocketRequest(req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "invalid authorization format")
}

// Test authentication with missing authorization header
func (suite *WebSocketHandlerTestSuite) TestAuthenticateWebSocketRequest_MissingHeader() {
	// Create request without authorization header
	req := httptest.NewRequest(http.MethodGet, "/collaborate", nil)

	// Execute
	result, err := suite.handler.authenticateWebSocketRequest(req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no authorization header provided")
}

// Test authentication from header
func (suite *WebSocketHandlerTestSuite) TestAuthenticateWebSocketRequest_Header() {
	userID := uuid.New()
	token := "header-token"

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

	// Create request with token in header
	req := httptest.NewRequest(http.MethodGet, "/collaborate", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute
	result, err := suite.handler.authenticateWebSocketRequest(req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, result.ID)
	suite.mockJWTService.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

// Helper method to add URL parameters to request
func (suite *WebSocketHandlerTestSuite) addURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	return req
}

// Benchmark tests
func BenchmarkWebSocketHandlerAuthentication(b *testing.B) {
	hub := websocketPkg.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	mockJWTService := new(mockService.MockJWTService)
	mockUserService := new(mockService.MockUserService)
	mockProjService := new(mockService.MockProjectService)

	handler := NewWebSocketHandler(hub, mockJWTService, mockUserService, mockProjService)

	userID := uuid.New()
	token := "test-token"
	user := testutil.CreateTestUser()
	user.ID = userID

	claims := &services.CustomClaims{
		UserID: userID,
		Email:  user.Email,
	}

	// Setup mocks for all benchmark iterations
	mockJWTService.On("ValidateToken", token).Return(claims, nil).Maybe()
	mockUserService.On("GetUserByID", userID).Return(user, nil).Maybe()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/collaborate?token="+token, nil)
		_, err := handler.authenticateWebSocketRequest(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}