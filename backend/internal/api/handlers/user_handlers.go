package handlers

import (
	"errors"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/api/utils"
	"github.com/Bug-Bugger/ezmodel/internal/services"
)

type UserHandler struct {
	userService services.UserServiceInterface
}

func NewUserHandler(userService services.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse and validate request body
		var req dto.CreateUserRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Create user through service
		user, err := h.userService.CreateUser(req.Email, req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserAlreadyExists):
				responses.RespondWithError(w, http.StatusConflict, "User already exists")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			}
			return
		}

		// Create user response without password
		userResponse := dto.UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "User created successfully", userResponse)
	}
}

func (h *UserHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "id", "Invalid user ID")
		if !ok {
			return
		}

		var req dto.UpdateUserRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Empty update request
		if req.Username == nil && req.Email == nil {
			responses.RespondWithError(w, http.StatusBadRequest, "No fields to update provided")
			return
		}

		user, err := h.userService.UpdateUser(id, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "User not found")
			case errors.Is(err, services.ErrUserAlreadyExists):
				responses.RespondWithError(w, http.StatusConflict, "Email already in use")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
			}
			return
		}

		userResponse := dto.UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "User updated successfully", userResponse)
	}
}

func (h *UserHandler) UpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "id", "Invalid user ID")
		if !ok {
			return
		}

		var req dto.UpdatePasswordRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		err := h.userService.UpdatePassword(id, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "User not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Password updated successfully", nil)
	}
}

func (h *UserHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "id", "Invalid user ID")
		if !ok {
			return
		}

		user, err := h.userService.GetUserByID(id)
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				responses.RespondWithError(w, http.StatusNotFound, "User not found")
			} else {
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user")
			}
			return
		}

		userResponse := dto.UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "User retrieved successfully", userResponse)
	}
}

func (h *UserHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.userService.GetAllUsers()
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
			return
		}

		var userResponses []dto.UserResponse
		for _, user := range users {
			userResponses = append(userResponses, dto.UserResponse{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Users retrieved successfully", userResponses)
	}
}

func (h *UserHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := utils.ParseUUIDParamWithError(w, r, "id", "Invalid user ID")
		if !ok {
			return
		}

		if err := h.userService.DeleteUser(id); err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				responses.RespondWithError(w, http.StatusNotFound, "User not found")
			} else {
				responses.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "User deleted successfully", nil)
	}
}
