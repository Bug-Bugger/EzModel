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

type TableHandler struct {
	tableService services.TableServiceInterface
}

func NewTableHandler(tableService services.TableServiceInterface) *TableHandler {
	return &TableHandler{
		tableService: tableService,
	}
}

// Create handles table creation within a project
func (h *TableHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Parse request body
		var req dto.CreateTableRequest
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

		// Create table through service
		table, err := h.tableService.CreateTable(projectID, req.Name, req.PosX, req.PosY)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrProjectNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Project not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		tableResponse := dto.TableResponse{
			ID:        table.ID,
			ProjectID: table.ProjectID,
			Name:      table.Name,
			PosX:      table.PosX,
			PosY:      table.PosY,
			CreatedAt: table.CreatedAt,
			UpdatedAt: table.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "Table created successfully", tableResponse)
	}
}

// GetByID handles retrieving a specific table
func (h *TableHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Get table from service
		table, err := h.tableService.GetTableByID(tableID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		tableResponse := dto.TableResponse{
			ID:        table.ID,
			ProjectID: table.ProjectID,
			Name:      table.Name,
			PosX:      table.PosX,
			PosY:      table.PosY,
			CreatedAt: table.CreatedAt,
			UpdatedAt: table.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Table retrieved successfully", tableResponse)
	}
}

// GetByProjectID handles retrieving all tables for a project
func (h *TableHandler) GetByProjectID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from URL
		projectIDStr := chi.URLParam(r, "id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid project ID format")
			return
		}

		// Get tables from service
		tables, err := h.tableService.GetTablesByProjectID(projectID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var tableResponses []dto.TableResponse
		for _, table := range tables {
			tableResponses = append(tableResponses, dto.TableResponse{
				ID:        table.ID,
				ProjectID: table.ProjectID,
				Name:      table.Name,
				PosX:      table.PosX,
				PosY:      table.PosY,
				CreatedAt: table.CreatedAt,
				UpdatedAt: table.UpdatedAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Tables retrieved successfully", tableResponses)
	}
}

// Update handles table updates
func (h *TableHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Parse request body
		var req dto.UpdateTableRequest
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

		// Update table through service
		table, err := h.tableService.UpdateTable(tableID, &req)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid input")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		tableResponse := dto.TableResponse{
			ID:        table.ID,
			ProjectID: table.ProjectID,
			Name:      table.Name,
			PosX:      table.PosX,
			PosY:      table.PosY,
			CreatedAt: table.CreatedAt,
			UpdatedAt: table.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Table updated successfully", tableResponse)
	}
}

// UpdatePosition handles table position updates
func (h *TableHandler) UpdatePosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Parse request body
		var req dto.UpdateTablePositionRequest
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

		// Update table position through service
		if err := h.tableService.UpdateTablePosition(tableID, req.PosX, req.PosY); err != nil {
			switch {
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Table position updated successfully", nil)
	}
}

// Delete handles table deletion
func (h *TableHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
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

		// Delete table through service with authorization check
		if err := h.tableService.DeleteTable(tableID, userID); err != nil {
			switch {
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			case errors.Is(err, services.ErrForbidden):
				responses.RespondWithError(w, http.StatusForbidden, "You don't have permission to delete this table")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Table deleted successfully", nil)
	}
}