package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	mockService "github.com/Bug-Bugger/ezmodel/internal/mocks/service"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	mockUserService *mockService.MockUserService
	mockJWTService  *mockService.MockJWTService
	handler         *AuthHandler
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	suite.mockUserService = new(mockService.MockUserService)
	suite.mockJWTService = new(mockService.MockJWTService)
	suite.handler = NewAuthHandler(suite.mockUserService, suite.mockJWTService)
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

// Test Login - Success
func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	// Setup
	loginRequest := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := testutil.CreateTestUser()
	tokenPair := &services.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
	}

	suite.mockUserService.On("AuthenticateUser", loginRequest.Email, loginRequest.Password).
		Return(user, nil)
	suite.mockJWTService.On("GenerateTokenPair", user).Return(tokenPair, nil)
	suite.mockJWTService.On("GetAccessTokenExpiration").Return(15 * time.Minute)

	// Make request
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", loginRequest)
	w := httptest.NewRecorder()

	// Execute
	suite.handler.Login()(w, req)

	// Assert
	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Login successful")

	tokenResponse, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should be a token object")
	suite.Equal(tokenPair.AccessToken, tokenResponse["access_token"])
	suite.Equal(tokenPair.RefreshToken, tokenResponse["refresh_token"])
	suite.Equal("Bearer", tokenResponse["token_type"])
	suite.Equal(float64(900), tokenResponse["expires_in"]) // 15 minutes = 900 seconds

	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test Login - Invalid JSON
func (suite *AuthHandlerTestSuite) TestLogin_InvalidJSON() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", "invalid json")
	w := httptest.NewRecorder()

	suite.handler.Login()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test Login - Invalid Credentials
func (suite *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	loginRequest := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	suite.mockUserService.On("AuthenticateUser", loginRequest.Email, loginRequest.Password).
		Return(nil, services.ErrInvalidCredentials)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", loginRequest)
	w := httptest.NewRecorder()

	suite.handler.Login()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Wrong email or password")
	suite.mockUserService.AssertExpectations(suite.T())
}

// Test Login - Authentication Service Error
func (suite *AuthHandlerTestSuite) TestLogin_AuthenticationServiceError() {
	loginRequest := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockUserService.On("AuthenticateUser", loginRequest.Email, loginRequest.Password).
		Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", loginRequest)
	w := httptest.NewRecorder()

	suite.handler.Login()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Login failed")
	suite.mockUserService.AssertExpectations(suite.T())
}

// Test Login - JWT Generation Error
func (suite *AuthHandlerTestSuite) TestLogin_JWTGenerationError() {
	loginRequest := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := testutil.CreateTestUser()
	suite.mockUserService.On("AuthenticateUser", loginRequest.Email, loginRequest.Password).
		Return(user, nil)
	suite.mockJWTService.On("GenerateTokenPair", user).Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", loginRequest)
	w := httptest.NewRecorder()

	suite.handler.Login()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Failed to generate tokens")
	suite.mockUserService.AssertExpectations(suite.T())
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Success
func (suite *AuthHandlerTestSuite) TestRefreshToken_Success() {
	refreshRequest := dto.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	newTokenPair := &services.TokenPair{
		AccessToken:  "new_access_token_123",
		RefreshToken: "new_refresh_token_123",
	}

	suite.mockJWTService.On("RefreshTokens", refreshRequest.RefreshToken).Return(newTokenPair, nil)
	suite.mockJWTService.On("GetAccessTokenExpiration").Return(15 * time.Minute)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/refresh-token", refreshRequest)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Token refreshed successfully")

	tokenResponse, ok := response.Data.(map[string]any)
	suite.True(ok)
	suite.Equal(newTokenPair.AccessToken, tokenResponse["access_token"])
	suite.Equal(newTokenPair.RefreshToken, tokenResponse["refresh_token"])
	suite.Equal("Bearer", tokenResponse["token_type"])

	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Invalid JSON
func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidJSON() {
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/refresh-token", "invalid json")
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid request body")
}

// Test RefreshToken - Invalid Token
func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	refreshRequest := dto.RefreshTokenRequest{
		RefreshToken: "invalid_refresh_token",
	}

	suite.mockJWTService.On("RefreshTokens", refreshRequest.RefreshToken).
		Return(nil, services.ErrInvalidToken)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/refresh-token", refreshRequest)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Invalid or expired refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Expired Token
func (suite *AuthHandlerTestSuite) TestRefreshToken_ExpiredToken() {
	refreshRequest := dto.RefreshTokenRequest{
		RefreshToken: "expired_refresh_token",
	}

	suite.mockJWTService.On("RefreshTokens", refreshRequest.RefreshToken).
		Return(nil, services.ErrExpiredToken)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/refresh-token", refreshRequest)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Invalid or expired refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Service Error
func (suite *AuthHandlerTestSuite) TestRefreshToken_ServiceError() {
	refreshRequest := dto.RefreshTokenRequest{
		RefreshToken: "some_refresh_token",
	}

	suite.mockJWTService.On("RefreshTokens", refreshRequest.RefreshToken).
		Return(nil, assert.AnError)

	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/refresh-token", refreshRequest)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Failed to refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}
