package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateTeamAndAddPR(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "frontend", 6)

	w := CreateTestPR(t, env.PRHandler, "pr-1020", "Add feature", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)

	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	pr := response["pr"].(map[string]interface{})

	if pr["pull_request_id"] != "pr-1020" {
		t.Errorf("expected pr-1020, got %v", pr["pull_request_id"])
	}

	reviewers := pr["assigned_reviewers"].([]interface{})

	if len(reviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(reviewers))
	}

	authorID := "u20"

	for _, reviewer := range reviewers {
		if reviewer == authorID {
			t.Errorf("author %s should not be a reviewer, got reviewers: %v", authorID, reviewers)
		}
	}
}
