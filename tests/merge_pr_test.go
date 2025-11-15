package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMergePR(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "devops", 5)

	w := CreateTestPR(t, env.PRHandler, "pr-2000", "Feature X", "u31")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	mergePayload := `{"pull_request_id": "pr-2000"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(mergePayload))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	env.PRHandler.MergePR(w2, req)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var response map[string]interface{}
	err := json.NewDecoder(w2.Body).Decode(&response)

	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	pr := response["pr"].(map[string]interface{})

	if pr["status"] != "MERGED" {
		t.Errorf("expected status MERGED, got %v", pr["status"])
	}
}
