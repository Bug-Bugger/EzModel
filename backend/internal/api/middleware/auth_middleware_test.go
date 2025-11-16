package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockService "github.com/Bug-Bugger/ezmodel/internal/mocks/service"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	mockJWTService *mockService.MockJWTService
	middleware     *AuthMiddleware
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	suite.mockJWTService = new(mockService.MockJWTService)
	suite.middleware = NewAuthMiddleware(suite.mockJWTService)
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticate_WithCookie() {
	token := "cookie-token"
	userID := uuid.New()
	claims := &services.CustomClaims{UserID: userID}

	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		ctxUserID, ok := GetUserIDFromContext(r.Context())
		suite.True(ok)
		suite.Equal(userID.String(), ctxUserID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()

	suite.middleware.Authenticate(next).ServeHTTP(w, req)

	suite.True(nextCalled)
	suite.Equal(http.StatusOK, w.Code)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticate_HeaderFallback() {
	token := "header-token"
	userID := uuid.New()
	claims := &services.CustomClaims{UserID: userID}

	suite.mockJWTService.On("ValidateToken", token).Return(claims, nil)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.middleware.Authenticate(next).ServeHTTP(w, req)

	suite.True(nextCalled)
	suite.Equal(http.StatusOK, w.Code)
	suite.mockJWTService.AssertExpectations(suite.T())
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticate_MissingCredentials() {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	suite.middleware.Authenticate(next).ServeHTTP(w, req)

	suite.False(nextCalled)
	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "No authentication credentials provided")
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticate_InvalidHeaderFormat() {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token sometoken")
	w := httptest.NewRecorder()

	suite.middleware.Authenticate(next).ServeHTTP(w, req)

	suite.False(nextCalled)
	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Invalid authorization format")
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticate_ExpiredToken() {
	token := "expired-token"

	suite.mockJWTService.On("ValidateToken", token).Return(nil, services.ErrExpiredToken)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	suite.middleware.Authenticate(next).ServeHTTP(w, req)

	suite.False(nextCalled)
	testutil.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "Token has expired")
	suite.mockJWTService.AssertExpectations(suite.T())
}
