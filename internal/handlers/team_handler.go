package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/services"
	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

type TeamHandler struct {
    teamService *services.TeamService
}

func NewTeamHandler(teamService *services.TeamService) *TeamHandler {
    return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    ctx := r.Context()
    var team models.Team

    if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    createdTeam, err := h.teamService.CreateTeam(ctx, &team)

    if err != nil {
        switch err.Error() {
        case "TEAM_EXISTS":
            respondError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
        case "team_name cannot be empty", "team must have at least one member":
            respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
        case "NOT_FOUND":
            respondError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
        default:
            if len(err.Error()) > 8 && (err.Error()[:8] == "duplicate" || err.Error()[:6] == "member") {
                respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
            } else {
                respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
            }
        }
        return
    }

    respondJSON(w, http.StatusCreated, map[string]interface{}{"team": createdTeam})
}


func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
        return
    }

    ctx := r.Context()
    teamName := r.URL.Query().Get("team_name")

    if teamName == "" {
        respondError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name query parameter required")
        return
    }

    team, err := h.teamService.GetTeam(ctx, teamName)

    if err != nil {
        if err.Error() == "NOT_FOUND" {
            respondError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
        } else {
            respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        }
        return
    }

    respondJSON(w, http.StatusOK, team)
}
