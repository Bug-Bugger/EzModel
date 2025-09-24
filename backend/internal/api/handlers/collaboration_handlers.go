package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/services"
	"github.com/Bug-Bugger/ezmodel/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CollaborationHandler struct {
	collaborationService services.CollaborationSessionServiceInterface
}

func NewCollaborationHandler(collaborationService services.CollaborationSessionServiceInterface) *CollaborationHandler {
	return &CollaborationHandler{
		collaborationService: collaborationService,
	}
}

// Create handles collaboration session creation within a project
func (h *CollaborationHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Get current user ID from context
		userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		// Parse request body
		var req dto.CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := validation.Validate(req); err != nil {
			validationErrors := validation.ValidationErrors(err)
			responses.RespondWithValidationErrors(w, validationErrors)
			return
		}

		// Create session through service
		session, err := h.collaborationService.CreateSession(projectID, userID, req.UserColor)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrUserNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "User not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		sessionResponse := dto.CollaborationSessionResponse{
			ID:         session.ID,
			ProjectID:  session.ProjectID,
			UserID:     session.UserID,
			CursorX:    session.CursorX,
			CursorY:    session.CursorY,
			UserColor:  session.UserColor,
			IsActive:   session.IsActive,
			LastPingAt: session.LastPingAt,
			JoinedAt:   session.JoinedAt,
			LeftAt:     session.LeftAt,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "Collaboration session created successfully", sessionResponse)
	}
}

// GetByID handles retrieving a specific collaboration session
func (h *CollaborationHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from URL
		sessionIDStr := chi.URLParam(r, "session_id")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid session ID format")
			return
		}

		// Get session from service
		session, err := h.collaborationService.GetSessionByID(sessionID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Collaboration session not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		sessionResponse := dto.CollaborationSessionResponse{
			ID:         session.ID,
			ProjectID:  session.ProjectID,
			UserID:     session.UserID,
			CursorX:    session.CursorX,
			CursorY:    session.CursorY,
			UserColor:  session.UserColor,
			IsActive:   session.IsActive,
			LastPingAt: session.LastPingAt,
			JoinedAt:   session.JoinedAt,
			LeftAt:     session.LeftAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaboration session retrieved successfully", sessionResponse)
	}
}

// GetByProjectID handles retrieving all collaboration sessions for a project
func (h *CollaborationHandler) GetByProjectID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Get sessions from service
		sessions, err := h.collaborationService.GetSessionsByProjectID(projectID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var sessionResponses []dto.CollaborationSessionResponse
		for _, session := range sessions {
			sessionResponses = append(sessionResponses, dto.CollaborationSessionResponse{
				ID:         session.ID,
				ProjectID:  session.ProjectID,
				UserID:     session.UserID,
				CursorX:    session.CursorX,
				CursorY:    session.CursorY,
				UserColor:  session.UserColor,
				IsActive:   session.IsActive,
				LastPingAt: session.LastPingAt,
				JoinedAt:   session.JoinedAt,
				LeftAt:     session.LeftAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaboration sessions retrieved successfully", sessionResponses)
	}
}

// GetActiveByProjectID handles retrieving active collaboration sessions for a project
func (h *CollaborationHandler) GetActiveByProjectID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Get active sessions from service
		sessions, err := h.collaborationService.GetActiveSessionsByProjectID(projectID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var sessionResponses []dto.CollaborationSessionResponse
		for _, session := range sessions {
			sessionResponses = append(sessionResponses, dto.CollaborationSessionResponse{
				ID:         session.ID,
				ProjectID:  session.ProjectID,
				UserID:     session.UserID,
				CursorX:    session.CursorX,
				CursorY:    session.CursorY,
				UserColor:  session.UserColor,
				IsActive:   session.IsActive,
				LastPingAt: session.LastPingAt,
				JoinedAt:   session.JoinedAt,
				LeftAt:     session.LeftAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Active collaboration sessions retrieved successfully", sessionResponses)
	}
}

// UpdateCursor handles cursor position updates
func (h *CollaborationHandler) UpdateCursor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from URL
		sessionIDStr := chi.URLParam(r, "session_id")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid session ID format")
			return
		}

		// Parse request body
		var req dto.UpdateCursorRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := validation.Validate(req); err != nil {
			validationErrors := validation.ValidationErrors(err)
			responses.RespondWithValidationErrors(w, validationErrors)
			return
		}

		// Update cursor through service
		if err := h.collaborationService.UpdateCursor(sessionID, req.CursorX, req.CursorY); err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Collaboration session not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Cursor position updated successfully", nil)
	}
}

// Update handles collaboration session updates
func (h *CollaborationHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from URL
		sessionIDStr := chi.URLParam(r, "session_id")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid session ID format")
			return
		}

		// Parse request body
		var req dto.UpdateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate input
		if err := validation.Validate(req); err != nil {
			validationErrors := validation.ValidationErrors(err)
			responses.RespondWithValidationErrors(w, validationErrors)
			return
		}

		// Update session through service
		session, err := h.collaborationService.UpdateSession(sessionID, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Collaboration session not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		sessionResponse := dto.CollaborationSessionResponse{
			ID:         session.ID,
			ProjectID:  session.ProjectID,
			UserID:     session.UserID,
			CursorX:    session.CursorX,
			CursorY:    session.CursorY,
			UserColor:  session.UserColor,
			IsActive:   session.IsActive,
			LastPingAt: session.LastPingAt,
			JoinedAt:   session.JoinedAt,
			LeftAt:     session.LeftAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaboration session updated successfully", sessionResponse)
	}
}

// SetInactive handles setting a collaboration session as inactive
func (h *CollaborationHandler) SetInactive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from URL
		sessionIDStr := chi.URLParam(r, "session_id")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid session ID format")
			return
		}

		// Set session inactive through service
		if err := h.collaborationService.SetSessionInactive(sessionID); err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Collaboration session not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaboration session set to inactive successfully", nil)
	}
}

// Delete handles collaboration session deletion
func (h *CollaborationHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from URL
		sessionIDStr := chi.URLParam(r, "session_id")
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid session ID format")
			return
		}

		// Get current user ID from context for authorization
		userIDStr, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}

		// Delete session through service with authorization check
		if err := h.collaborationService.DeleteSession(sessionID, userID); err != nil {
			switch {
			case errors.Is(err, services.ErrSessionNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Collaboration session not found")
			case errors.Is(err, services.ErrForbidden):
				responses.RespondWithError(w, http.StatusForbidden, "You don't have permission to delete this collaboration session")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Collaboration session deleted successfully", nil)
	}
}