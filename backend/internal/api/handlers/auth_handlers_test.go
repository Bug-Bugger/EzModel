package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/config"
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
	cfg             *config.Config
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	suite.mockUserService = new(mockService.MockUserService)
	suite.mockJWTService = new(mockService.MockJWTService)
	suite.cfg = config.New()
	suite.handler = NewAuthHandler(suite.mockUserService, suite.mockJWTService, suite.cfg)
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (suite *AuthHandlerTestSuite) getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
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
	suite.mockJWTService.On("GetRefreshTokenExpiration").Return(7 * 24 * time.Hour)

	// Make request
	req := testutil.MakeJSONRequest(suite.T(), http.MethodPost, "/login", loginRequest)
	w := httptest.NewRecorder()

	// Execute
	suite.handler.Login()(w, req)

	// Assert
	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Login successful")
	userPayload, ok := response.Data.(map[string]any)
	suite.True(ok, "Response data should contain user payload")
	userData, ok := userPayload["user"].(map[string]any)
	suite.True(ok, "User payload should be a map")
	suite.Equal(user.Email, userData["email"])
	suite.Equal(user.Username, userData["username"])

	result := w.Result()
	accessCookie := suite.getCookie(result.Cookies(), "access_token")
	refreshCookie := suite.getCookie(result.Cookies(), "refresh_token")

	suite.NotNil(accessCookie)
	suite.Equal(tokenPair.AccessToken, accessCookie.Value)
	suite.True(accessCookie.HttpOnly)
	suite.Equal(suite.cfg.Env == "production", accessCookie.Secure)
	suite.Equal(http.SameSiteStrictMode, accessCookie.SameSite)
	suite.Equal("/", accessCookie.Path)
	suite.Equal(900, accessCookie.MaxAge)

	suite.NotNil(refreshCookie)
	suite.Equal(tokenPair.RefreshToken, refreshCookie.Value)
	suite.True(refreshCookie.HttpOnly)
	suite.Equal(suite.cfg.Env == "production", refreshCookie.Secure)
	suite.Equal(http.SameSiteStrictMode, refreshCookie.SameSite)
	suite.Equal("/", refreshCookie.Path)
	suite.Equal(7*24*60*60, refreshCookie.MaxAge)

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
	refreshToken := "valid_refresh_token"
	newTokenPair := &services.TokenPair{
		AccessToken:  "new_access_token_123",
		RefreshToken: "new_refresh_token_123",
	}

	suite.mockJWTService.On("RefreshTokens", refreshToken).Return(newTokenPair, nil)
	suite.mockJWTService.On("GetAccessTokenExpiration").Return(15 * time.Minute)
	suite.mockJWTService.On("GetRefreshTokenExpiration").Return(7 * 24 * time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Token refreshed successfully")
	suite.Nil(response.Data)

	result := w.Result()
	accessCookie := suite.getCookie(result.Cookies(), "access_token")
	refreshCookie := suite.getCookie(result.Cookies(), "refresh_token")

	suite.NotNil(accessCookie)
	suite.Equal(newTokenPair.AccessToken, accessCookie.Value)
	suite.Equal(900, accessCookie.MaxAge)
	suite.NotNil(refreshCookie)
	suite.Equal(newTokenPair.RefreshToken, refreshCookie.Value)
	suite.Equal(7*24*60*60, refreshCookie.MaxAge)

	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestRefreshToken_MissingCookie() {
	req := httptest.NewRequest(http.MethodPost, "/refresh-token", nil)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "No refresh token found")
}

// Test RefreshToken - Invalid Token
func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	refreshToken := "invalid_refresh_token"
	suite.mockJWTService.On("RefreshTokens", refreshToken).
		Return(nil, services.ErrInvalidToken)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Invalid or expired refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Expired Token
func (suite *AuthHandlerTestSuite) TestRefreshToken_ExpiredToken() {
	refreshToken := "expired_refresh_token"
	suite.mockJWTService.On("RefreshTokens", refreshToken).
		Return(nil, services.ErrExpiredToken)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Invalid or expired refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

// Test RefreshToken - Service Error
func (suite *AuthHandlerTestSuite) TestRefreshToken_ServiceError() {
	refreshToken := "some_refresh_token"
	suite.mockJWTService.On("RefreshTokens", refreshToken).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	w := httptest.NewRecorder()

	suite.handler.RefreshToken()(w, req)

	testutil.AssertErrorResponse(suite.T(), w, http.StatusInternalServerError, "Failed to refresh token")
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogout_Success() {
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	suite.handler.Logout()(w, req)

	response := testutil.AssertSuccessResponse(suite.T(), w, http.StatusOK, "Logout successful")
	suite.Nil(response.Data)

	result := w.Result()
	accessCookie := suite.getCookie(result.Cookies(), "access_token")
	refreshCookie := suite.getCookie(result.Cookies(), "refresh_token")

	suite.NotNil(accessCookie)
	suite.Equal(-1, accessCookie.MaxAge)
	suite.Equal("", accessCookie.Value)

	suite.NotNil(refreshCookie)
	suite.Equal(-1, refreshCookie.MaxAge)
	suite.Equal("", refreshCookie.Value)
}
