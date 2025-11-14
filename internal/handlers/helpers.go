package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Error: models.ErrorDetail{Code: code, Message: message},
	})
}