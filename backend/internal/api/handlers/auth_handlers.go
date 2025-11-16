package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/config"
	"github.com/Bug-Bugger/ezmodel/internal/services"
)

type AuthHandler struct {
	userService services.UserServiceInterface
	jwtService  services.JWTServiceInterface
	cfg         *config.Config
}

func NewAuthHandler(userService services.UserServiceInterface, jwtService services.JWTServiceInterface, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
		cfg:         cfg,
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

		// Set access token as httpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    tokens.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production", // Only send over HTTPS in production
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(h.jwtService.GetAccessTokenExpiration() / time.Second),
		})

		// Set refresh token as httpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    tokens.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production", // Only send over HTTPS in production
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(h.jwtService.GetRefreshTokenExpiration() / time.Second),
		})

		// Return user data without tokens
		userResponse := map[string]interface{}{
			"user": user,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Login successful", userResponse)
	}
}

func (h *AuthHandler) RefreshToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			responses.RespondWithError(w, http.StatusUnauthorized, "No refresh token found")
			return
		}

		refreshToken := cookie.Value

		tokens, err := h.jwtService.RefreshTokens(refreshToken)
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

		// Set new access token as httpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    tokens.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(h.jwtService.GetAccessTokenExpiration() / time.Second),
		})

		// Set new refresh token as httpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    tokens.RefreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(h.jwtService.GetRefreshTokenExpiration() / time.Second),
		})

		responses.RespondWithSuccess(w, http.StatusOK, "Token refreshed successfully", nil)
	}
}

func (h *AuthHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear access token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1, // Delete cookie
		})

		// Clear refresh token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1, // Delete cookie
		})

		responses.RespondWithSuccess(w, http.StatusOK, "Logout successful", nil)
	}
}
