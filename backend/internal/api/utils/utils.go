package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/responses"
	"github.com/Bug-Bugger/ezmodel/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ParseUUIDParam extracts and parses a UUID parameter from the URL path
func ParseUUIDParam(w http.ResponseWriter, r *http.Request, paramName string) (uuid.UUID, bool) {
	paramStr := chi.URLParam(r, paramName)
	paramUUID, err := uuid.Parse(paramStr)
	if err != nil {
		var errorMessage string
		switch paramName {
		case "table_id":
			errorMessage = "Invalid table ID format"
		case "field_id":
			errorMessage = "Invalid field ID format"
		case "user_id":
			errorMessage = "Invalid user ID format"
		case "relationship_id":
			errorMessage = "Invalid relationship ID format"
		case "collaborator_id":
			errorMessage = "Invalid collaborator ID"
		case "session_id":
			errorMessage = "Invalid session ID format"
		default:
			errorMessage = "Invalid " + paramName + " format"
		}
		responses.RespondWithError(w, http.StatusBadRequest, errorMessage)
		return uuid.Nil, false
	}
	return paramUUID, true
}

// ParseUUIDParamWithError extracts and parses a UUID parameter from the URL path with custom error message
func ParseUUIDParamWithError(w http.ResponseWriter, r *http.Request, paramName, errorMessage string) (uuid.UUID, bool) {
	paramStr := chi.URLParam(r, paramName)
	paramUUID, err := uuid.Parse(paramStr)
	if err != nil {
		responses.RespondWithError(w, http.StatusBadRequest, errorMessage)
		return uuid.Nil, false
	}
	return paramUUID, true
}

// DecodeAndValidate decodes JSON request body into the provided struct and validates it
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, requestStruct any) bool {
	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(requestStruct); err != nil {
		responses.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}

	// Validate input
	if err := validation.Validate(requestStruct); err != nil {
		validationErrors := validation.ValidationErrors(err)
		responses.RespondWithValidationErrors(w, validationErrors)
		return false
	}

	return true
}
