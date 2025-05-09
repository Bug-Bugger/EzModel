package responses

import (
	"encoding/json"
	"net/http"

	"github.com/Bug-Bugger/ezmodel/internal/api/dto"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, dto.APIResponse{
		Success: false,
		Message: message,
	})
}

func RespondWithValidationErrors(w http.ResponseWriter, errors map[string]string) {
	respondWithJSON(w, http.StatusBadRequest, dto.APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	respondWithJSON(w, code, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
