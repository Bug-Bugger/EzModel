package handlers

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

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, dto.APIResponse{
		Success: false,
		Message: message,
	})
}

func respondWithValidationErrors(w http.ResponseWriter, errors map[string]string) {
	respondWithJSON(w, http.StatusBadRequest, dto.APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

func respondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	respondWithJSON(w, code, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
