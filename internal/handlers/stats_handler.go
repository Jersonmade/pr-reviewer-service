package handlers

import (
	"net/http"

	"github.com/Jersonmade/pr-reviewer-service/internal/services"
)

type AnalyticsHandler struct {
	analyticsService *services.StatsService
}

func NewAnalyticsHandler(analyticsService *services.StatsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) GetReviewAssignmentsStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET allowed")
		return
	}

	ctx := r.Context()
	counts, err := h.analyticsService.GetReviewAssignmentsCount(ctx)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"review_assignments": counts,
	})
}
