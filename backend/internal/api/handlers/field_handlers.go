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

type FieldHandler struct {
	fieldService services.FieldServiceInterface
}

func NewFieldHandler(fieldService services.FieldServiceInterface) *FieldHandler {
	return &FieldHandler{
		fieldService: fieldService,
	}
}

// Create handles field creation within a table
func (h *FieldHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Parse request body
		var req dto.CreateFieldRequest
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

		// Create field through service
		field, err := h.fieldService.CreateField(tableID, &req)
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
		fieldResponse := dto.FieldResponse{
			ID:           field.ID,
			TableID:      field.TableID,
			Name:         field.Name,
			DataType:     field.DataType,
			IsPrimaryKey: field.IsPrimaryKey,
			IsNullable:   field.IsNullable,
			DefaultValue: field.DefaultValue,
			Position:     field.Position,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusCreated, "Field created successfully", fieldResponse)
	}
}

// GetByID handles retrieving a specific field
func (h *FieldHandler) GetByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get field ID from URL
		fieldIDStr := chi.URLParam(r, "field_id")
		fieldID, err := uuid.Parse(fieldIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid field ID format")
			return
		}

		// Get field from service
		field, err := h.fieldService.GetFieldByID(fieldID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrFieldNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Field not found")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Convert to response format
		fieldResponse := dto.FieldResponse{
			ID:           field.ID,
			TableID:      field.TableID,
			Name:         field.Name,
			DataType:     field.DataType,
			IsPrimaryKey: field.IsPrimaryKey,
			IsNullable:   field.IsNullable,
			DefaultValue: field.DefaultValue,
			Position:     field.Position,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Field retrieved successfully", fieldResponse)
	}
}

// GetByTableID handles retrieving all fields for a table
func (h *FieldHandler) GetByTableID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Get fields from service
		fields, err := h.fieldService.GetFieldsByTableID(tableID)
		if err != nil {
			responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Convert to response format
		var fieldResponses []dto.FieldResponse
		for _, field := range fields {
			fieldResponses = append(fieldResponses, dto.FieldResponse{
				ID:           field.ID,
				TableID:      field.TableID,
				Name:         field.Name,
				DataType:     field.DataType,
				IsPrimaryKey: field.IsPrimaryKey,
				IsNullable:   field.IsNullable,
				DefaultValue: field.DefaultValue,
				Position:     field.Position,
				CreatedAt:    field.CreatedAt,
				UpdatedAt:    field.UpdatedAt,
			})
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Fields retrieved successfully", fieldResponses)
	}
}

// Update handles field updates
func (h *FieldHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get field ID from URL
		fieldIDStr := chi.URLParam(r, "field_id")
		fieldID, err := uuid.Parse(fieldIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid field ID format")
			return
		}

		// Parse request body
		var req dto.UpdateFieldRequest
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

		// Update field through service
		field, err := h.fieldService.UpdateField(fieldID, &req)
		if err != nil {
			switch {
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
		fieldResponse := dto.FieldResponse{
			ID:           field.ID,
			TableID:      field.TableID,
			Name:         field.Name,
			DataType:     field.DataType,
			IsPrimaryKey: field.IsPrimaryKey,
			IsNullable:   field.IsNullable,
			DefaultValue: field.DefaultValue,
			Position:     field.Position,
			CreatedAt:    field.CreatedAt,
			UpdatedAt:    field.UpdatedAt,
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Field updated successfully", fieldResponse)
	}
}

// Reorder handles field reordering within a table
func (h *FieldHandler) Reorder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get table ID from URL
		tableIDStr := chi.URLParam(r, "table_id")
		tableID, err := uuid.Parse(tableIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid table ID format")
			return
		}

		// Parse request body
		var req dto.ReorderFieldsRequest
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

		// Reorder fields through service
		if err := h.fieldService.ReorderFields(tableID, req.FieldPositions); err != nil {
			switch {
			case errors.Is(err, services.ErrTableNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Table not found")
			case errors.Is(err, services.ErrFieldNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "One or more fields not found")
			case errors.Is(err, services.ErrInvalidInput):
				responses.RespondWithError(w, http.StatusBadRequest, "Invalid field positions")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Fields reordered successfully", nil)
	}
}

// Delete handles field deletion
func (h *FieldHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get field ID from URL
		fieldIDStr := chi.URLParam(r, "field_id")
		fieldID, err := uuid.Parse(fieldIDStr)
		if err != nil {
			responses.RespondWithError(w, http.StatusBadRequest, "Invalid field ID format")
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

		// Delete field through service with authorization check
		if err := h.fieldService.DeleteField(fieldID, userID); err != nil {
			switch {
			case errors.Is(err, services.ErrFieldNotFound):
				responses.RespondWithError(w, http.StatusNotFound, "Field not found")
			case errors.Is(err, services.ErrForbidden):
				responses.RespondWithError(w, http.StatusForbidden, "You don't have permission to delete this field")
			default:
				responses.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		responses.RespondWithSuccess(w, http.StatusOK, "Field deleted successfully", nil)
	}
}