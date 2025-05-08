package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/services"
)

type AuthHandler struct {
	userService services.UserServiceInterface
	jwtService  *services.JWTService
}

func NewAuthHandler(userService services.UserServiceInterface, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		user, err := h.userService.AuthenticateUser(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				responses.RespondWithError(w, http.StatusUnauthorized, "Wrong email or password")
			} else {
				responses.RespondWithError(w, http.StatusInternalServerError, "Login failed")
			}
			return
		}

		// Generate JWT tokens
		tokens, err := h.jwtService.GenerateTokenPair(user)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
			return
		}

		expiresInSeconds := int(h.jwtService.GetAccessTokenExpiration() / time.Second)
		tokenResponse := dto.TokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expiresInSeconds,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Login successful", tokenResponse)
	}
}

func (h *AuthHandler) RefreshToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		tokens, err := h.jwtService.RefreshTokens(req.RefreshToken)
		if err != nil {
			statusCode := http.StatusInternalServerError
			message := "Failed to refresh token"

			if errors.Is(err, services.ErrInvalidToken) || errors.Is(err, services.ErrExpiredToken) {
				statusCode = http.StatusUnauthorized
				message = "Invalid or expired refresh token"
			}

			responses.RespondWithError(w, statusCode, message)
			return
		}

		expiresInSeconds := int(h.jwtService.GetAccessTokenExpiration() / time.Second)
		tokenResponse := dto.TokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expiresInSeconds,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Token refreshed successfully", tokenResponse)
	}
}
