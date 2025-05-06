package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/services"
)

type AuthHandler struct {
	userService services.UserServiceInterface
}

func NewAuthHandler(userService services.UserServiceInterface) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		user, err := h.userService.AuthenticateUser(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				respondWithError(w, http.StatusUnauthorized, "Wrong email or password")
			} else {
				respondWithError(w, http.StatusInternalServerError, "Login failed")
			}
			return
		}

		userResponse := dto.UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Username:      user.Username,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
		}

		respondWithSuccess(w, http.StatusOK, "Login successful", userResponse)
	}
}
