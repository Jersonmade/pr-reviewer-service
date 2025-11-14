package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/Jersonmade/pr-reviewer-service/internal/services"
)

type PRHandler struct {
    prService *services.PRService
}

func NewPRHandler(prService *services.PRService) *PRHandler {
    return &PRHandler{prService: prService}
}


func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    ctx := r.Context()

    var req struct {
        PullRequestID   string `json:"pull_request_id"`
        PullRequestName string `json:"pull_request_name"`
        AuthorID        string `json:"author_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    pr, err := h.prService.CreatePR(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)

    if err != nil {
        switch err.Error() {
        case "AUTHOR_NOT_FOUND", "USER_NOT_FOUND":
            respondError(w, http.StatusNotFound, "NOT_FOUND", "author not found")
        case "PR_EXISTS":
            respondError(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
        case "pull_request_id cannot be empty", "pull_request_name cannot be empty", "author_id cannot be empty":
            respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
        default:
            respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        }
        return
    }

    respondJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}


func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    ctx := r.Context()

    var req struct {
        PullRequestID string `json:"pull_request_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    pr, err := h.prService.MergePR(ctx, req.PullRequestID)

    if err != nil {
        if err.Error() == "PR_NOT_FOUND" {
            respondError(w, http.StatusNotFound, "NOT_FOUND", "PR not found")
        } else {
            respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        }

        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}


func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        respondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    ctx := r.Context()

    var req struct {
        PullRequestID string `json:"pull_request_id"`
        OldUserID     string `json:"old_user_id"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    newReviewerID, err := h.prService.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
    
    if err != nil {
        switch err.Error() {
        case "PR_NOT_FOUND", "USER_NOT_FOUND":
            respondError(w, http.StatusNotFound, "NOT_FOUND", "PR or user not found")
        case "PR_MERGED":
            respondError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
        case "NOT_ASSIGNED":
            respondError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
        case "NO_CANDIDATE":
            respondError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
        default:
            respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
        }
        return
    }

    pr, _ := h.prService.GetPR(ctx, req.PullRequestID)

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "pr":          pr,
        "replaced_by": newReviewerID,
    })
}
