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

type RelationshipHandler struct {
	relationshipService services.RelationshipServiceInterface
}

func NewRelationshipHandler(relationshipService services.RelationshipServiceInterface) *RelationshipHandler {
	return &RelationshipHandler{
		relationshipService: relationshipService,
	}
}

// Create handles relationship creation within a project
func (h *RelationshipHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Parse request body
		var req dto.CreateRelationshipRequest
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

		// Create relationship through service
		relationship, err := h.relationshipService.CreateRelationship(projectID, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			case errors.Is(err, services.ErrFieldNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Field not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		relationshipResponse := dto.RelationshipResponse{
			ID:            relationship.ID,
			ProjectID:     relationship.ProjectID,
			SourceTableID: relationship.SourceTableID,
			SourceFieldID: relationship.SourceFieldID,
			TargetTableID: relationship.TargetTableID,
			TargetFieldID: relationship.TargetFieldID,
			RelationType:  relationship.RelationType,
			CreatedAt:     relationship.CreatedAt,
			UpdatedAt:     relationship.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "Relationship created successfully", relationshipResponse)
	}
}

// GetByID handles retrieving a specific relationship
func (h *RelationshipHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get relationship ID from URL
		relationshipIDStr := chi.URLParam(r, "relationship_id")
		relationshipID, err := uuid.Parse(relationshipIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid relationship ID format")
			return
		}

		// Get relationship from service
		relationship, err := h.relationshipService.GetRelationshipByID(relationshipID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrRelationshipNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Relationship not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		relationshipResponse := dto.RelationshipResponse{
			ID:            relationship.ID,
			ProjectID:     relationship.ProjectID,
			SourceTableID: relationship.SourceTableID,
			SourceFieldID: relationship.SourceFieldID,
			TargetTableID: relationship.TargetTableID,
			TargetFieldID: relationship.TargetFieldID,
			RelationType:  relationship.RelationType,
			CreatedAt:     relationship.CreatedAt,
			UpdatedAt:     relationship.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationship retrieved successfully", relationshipResponse)
	}
}

// GetByProjectID handles retrieving all relationships for a project
func (h *RelationshipHandler) GetByProjectID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Get relationships from service
		relationships, err := h.relationshipService.GetRelationshipsByProjectID(projectID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var relationshipResponses []dto.RelationshipResponse
		for _, relationship := range relationships {
			relationshipResponses = append(relationshipResponses, dto.RelationshipResponse{
				ID:            relationship.ID,
				ProjectID:     relationship.ProjectID,
				SourceTableID: relationship.SourceTableID,
				SourceFieldID: relationship.SourceFieldID,
				TargetTableID: relationship.TargetTableID,
				TargetFieldID: relationship.TargetFieldID,
				RelationType:  relationship.RelationType,
				CreatedAt:     relationship.CreatedAt,
				UpdatedAt:     relationship.UpdatedAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationships retrieved successfully", relationshipResponses)
	}
}

// GetByTableID handles retrieving all relationships for a table
func (h *RelationshipHandler) GetByTableID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Get relationships from service
		relationships, err := h.relationshipService.GetRelationshipsByTableID(tableID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var relationshipResponses []dto.RelationshipResponse
		for _, relationship := range relationships {
			relationshipResponses = append(relationshipResponses, dto.RelationshipResponse{
				ID:            relationship.ID,
				ProjectID:     relationship.ProjectID,
				SourceTableID: relationship.SourceTableID,
				SourceFieldID: relationship.SourceFieldID,
				TargetTableID: relationship.TargetTableID,
				TargetFieldID: relationship.TargetFieldID,
				RelationType:  relationship.RelationType,
				CreatedAt:     relationship.CreatedAt,
				UpdatedAt:     relationship.UpdatedAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationships retrieved successfully", relationshipResponses)
	}
}

// Update handles relationship updates
func (h *RelationshipHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get relationship ID from URL
		relationshipIDStr := chi.URLParam(r, "relationship_id")
		relationshipID, err := uuid.Parse(relationshipIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid relationship ID format")
			return
		}

		// Parse request body
		var req dto.UpdateRelationshipRequest
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

		// Update relationship through service
		relationship, err := h.relationshipService.UpdateRelationship(relationshipID, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrRelationshipNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Relationship not found")
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			case errors.Is(err, services.ErrFieldNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Field not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		relationshipResponse := dto.RelationshipResponse{
			ID:            relationship.ID,
			ProjectID:     relationship.ProjectID,
			SourceTableID: relationship.SourceTableID,
			SourceFieldID: relationship.SourceFieldID,
			TargetTableID: relationship.TargetTableID,
			TargetFieldID: relationship.TargetFieldID,
			RelationType:  relationship.RelationType,
			CreatedAt:     relationship.CreatedAt,
			UpdatedAt:     relationship.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationship updated successfully", relationshipResponse)
	}
}

// Delete handles relationship deletion
func (h *RelationshipHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get relationship ID from URL
		relationshipIDStr := chi.URLParam(r, "relationship_id")
		relationshipID, err := uuid.Parse(relationshipIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid relationship ID format")
			return
		}

		// Get current user ID from context for authorization
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			responses.RespondWithError(w, http.StatusUnauthorized, "User context not found")
			return
		}

		// TODO: Add authorization check to ensure user can delete this relationship
		_ = userID // Use userID for authorization logic

		// Delete relationship through service
		if err := h.relationshipService.DeleteRelationship(relationshipID); err != nil {
			switch {
			case errors.Is(err, services.ErrRelationshipNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Relationship not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationship deleted successfully", nil)
	}
}