package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
	"github.com/Jersonmade/pr-reviewer-service/internal/services"
)

type TeamHandler struct {
	teamService *services.TeamService
}

func NewTeamHandler(teamService *services.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	ctx := r.Context()
	var team models.Team

	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	createdTeam, err := h.teamService.CreateTeam(ctx, &team)

	if err != nil {
		switch err.Error() {
		case "TEAM_EXISTS":
			RespondError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
			return
		case "team_name cannot be empty", "team must have at least one member":
			RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		case "NOT_FOUND":
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
			return
		default:
			if strings.HasPrefix(err.Error(), "duplicate user_id") ||
				strings.HasPrefix(err.Error(), "member at index") {
				RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
				return
			}
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{"team": createdTeam})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
		return
	}

	ctx := r.Context()
	teamName := r.URL.Query().Get("team_name")

	if teamName == "" {
		RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name query parameter required")
		return
	}

	team, err := h.teamService.GetTeam(ctx, teamName)

	if err != nil {
		if err.Error() == "NOT_FOUND" {
			RespondError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
		} else {
			RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, team)
}
