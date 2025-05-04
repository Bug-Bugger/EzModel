package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/validation"
	"github.com/go-chi/chi/v5"
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
		// Parse request body
		var req dto.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := validation.Validate(req); err != nil {
			validationErrors := validation.ValidationErrors(err)
			respondWithValidationErrors(w, validationErrors)
			return
		}

		// Create user through service
		user, err := h.userService.CreateUser(req.Name)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserAlreadyExists):
				respondWithError(w, http.StatusConflict, "User already exists")
			case errors.Is(err, services.ErrInvalidInput):
				respondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				respondWithError(w, http.StatusInternalServerError, "Failed to create user")
			}
			return
		}

		respondWithSuccess(w, http.StatusCreated, "User created successfully", user)
	}
}

func (h *UserHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var req dto.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := validation.Validate(req); err != nil {
			validationErrors := validation.ValidationErrors(err)
			respondWithValidationErrors(w, validationErrors)
			return
		}

		user, err := h.userService.UpdateUser(id, req.Name)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				respondWithError(w, http.StatusNotFound, "User not found")
			case errors.Is(err, services.ErrInvalidInput):
				respondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				respondWithError(w, http.StatusInternalServerError, "Failed to update user")
			}
			return
		}

		respondWithSuccess(w, http.StatusOK, "User updated successfully", user)
	}
}

func (h *UserHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		user, err := h.userService.GetUserByID(id)
		if err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				respondWithError(w, http.StatusNotFound, "User not found")
			} else {
				respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user")
			}
			return
		}

		respondWithSuccess(w, http.StatusOK, "User retrieved successfully", user)
	}
}

func (h *UserHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.userService.GetAllUsers()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
			return
		}

		respondWithSuccess(w, http.StatusOK, "Users retrieved successfully", users)
	}
}

func (h *UserHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		if err := h.userService.DeleteUser(id); err != nil {
			if errors.Is(err, services.ErrUserNotFound) {
				respondWithError(w, http.StatusNotFound, "User not found")
			} else {
				respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
