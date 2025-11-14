package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Error: models.ErrorDetail{Code: code, Message: message},
	})
}
