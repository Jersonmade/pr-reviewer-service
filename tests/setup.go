package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Jersonmade/pr-reviewer-service/internal/handlers"
)

func CreateTestTeam(t *testing.T, handler *handlers.TeamHandler, teamName string, memberCount int) {
	t.Helper()

	members := make([]map[string]interface{}, memberCount)
	for i := 0; i < memberCount; i++ {
		members[i] = map[string]interface{}{
			"user_id":   fmt.Sprintf("u%d", i+30),
			"username":  fmt.Sprintf("User%d", i+30),
			"is_active": true,
		}
	}

	payload := map[string]interface{}{
		"team_name": teamName,
		"members":   members,
	}

	jsonBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddTeam(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create team: %d - %s", w.Code, w.Body.String())
	}
}

func CreateTestPR(t *testing.T, handler *handlers.PRHandler, prID, prName, authorID string) *httptest.ResponseRecorder {
	t.Helper()

	payload := map[string]string{
		"pull_request_id":   prID,
		"pull_request_name": prName,
		"author_id":         authorID,
	}

	jsonBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePR(w, req)

	return w
}
