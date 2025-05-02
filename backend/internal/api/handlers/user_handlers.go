package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/Bug-Bugger/ezmodel/internal/repository"
	"github.com/Bug-Bugger/ezmodel/internal/validation"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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

		// Create user
		user := &models.User{
			Name: req.Name,
		}

		id, err := h.userRepo.Create(user)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		user.ID = id
		respondWithSuccess(w, http.StatusCreated, "User created successfully", user)
	}
}

func (h *UserHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/users/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		existingUser, err := h.userRepo.GetByID(id)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User not found")
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

		existingUser.Name = req.Name
		if err := h.userRepo.Update(existingUser); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		respondWithSuccess(w, http.StatusOK, "User updated successfully", existingUser)
	}
}

func (h *UserHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/users/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		user, err := h.userRepo.GetByID(id)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		respondWithSuccess(w, http.StatusOK, "User retrieved successfully", user)
	}
}

func (h *UserHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		users, err := h.userRepo.GetAll()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
			return
		}

		respondWithSuccess(w, http.StatusOK, "Users retrieved successfully", users)
	}
}

func (h *UserHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/users/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		if err := h.userRepo.Delete(id); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
