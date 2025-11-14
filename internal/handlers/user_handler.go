package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/services"
)

type UserHandler struct {
	userService *services.UserService
	prService   *services.PRService
}

func NewUserHandler(userService *services.UserService, prService *services.PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	ctx := r.Context()

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	user, err := h.userService.SetUserActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		switch err.Error() {
		case "USER_NOT_FOUND":
			RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		case "User ID cannot be empty":
			RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		default:
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
		return
	}

	ctx := r.Context()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "user_id query parameter required")
		return
	}

	_, err := h.userService.GetUser(ctx, userID)

	if err != nil {
		if err.Error() == "USER_NOT_FOUND" {
			RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		} else {
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	prs, err := h.prService.GetPRsByReviewer(ctx, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
