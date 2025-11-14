package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/services"
)

type UserHandler struct {
	userService *services.UserService
	prService *services.PRService
}


func NewUserHandler (userService *services.UserService, prService *services.PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService: prService,
	}
}


func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    ctx := r.Context()

    var req struct {
        UserID   string `json:"user_id"`
        IsActive bool   `json:"is_active"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    user, err := h.userService.SetUserActive(ctx, req.UserID, req.IsActive)
    if err != nil {
        if err.Error() == "NOT_FOUND" {
            respondError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
        } else {
            respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}


func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
        return
    }

    ctx := r.Context()
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        respondError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id query parameter required")
        return
    }

    if _, err := h.userService.GetUser(ctx, userID); err != nil {
        respondError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
        return
    }

    prs, err := h.prService.GetPRsByReviewer(ctx, userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "user_id":       userID,
        "pull_requests": prs,
    })
}
