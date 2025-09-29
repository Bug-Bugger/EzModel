package handlers

import (
	"errors"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
	"github.com/Bug-Bugger/ezmodel/internal/api/middleware"
	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/api/utils"
	"github.com/Bug-Bugger/ezmodel/internal/services"
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
		projectID, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID format")
		if !ok {
			return
		}

		// Parse and validate request body
		var req dto.CreateRelationshipRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Get current user ID from context for collaboration
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

		// Create relationship through service
		relationship, err := h.relationshipService.CreateRelationship(projectID, &req, userID)
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
		relationshipID, ok := utils.ParseUUIDParam(w, r, "relationship_id")
		if !ok {
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
		projectID, ok := utils.ParseUUIDParamWithError(w, r, "project_id", "Invalid project ID format")
		if !ok {
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
		tableID, ok := utils.ParseUUIDParam(w, r, "table_id")
		if !ok {
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
		relationshipID, ok := utils.ParseUUIDParam(w, r, "relationship_id")
		if !ok {
			return
		}

		// Parse and validate request body
		var req dto.UpdateRelationshipRequest
		if !utils.DecodeAndValidate(w, r, &req) {
			return
		}

		// Get current user ID from context for collaboration
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

		// Update relationship through service
		relationship, err := h.relationshipService.UpdateRelationship(relationshipID, &req, userID)
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
		relationshipID, ok := utils.ParseUUIDParam(w, r, "relationship_id")
		if !ok {
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

		// Delete relationship through service with authorization check
		if err := h.relationshipService.DeleteRelationship(relationshipID, userID); err != nil {
			switch {
			case errors.Is(err, services.ErrRelationshipNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Relationship not found")
			case errors.Is(err, services.ErrForbidden):
				responses.RespondWithError(w, http.StatusForbidden, "You don't have permission to delete this relationship")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Relationship deleted successfully", nil)
	}
}
