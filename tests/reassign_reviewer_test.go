package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReassignReviewer(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "backend", 5)

	w := CreateTestPR(t, env.PRHandler, "pr-3000", "Refactoring", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var createResponse map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&createResponse); err != nil {
		t.Fatalf("failed to decode create PR response: %v", err)
	}

	pr := createResponse["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	if len(reviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(reviewers))
	}

	oldReviewerID := reviewers[0].(string)

	reassignPayload := map[string]string{
		"pull_request_id": "pr-3000",
		"old_user_id":     oldReviewerID,
	}

	jsonBytes, _ := json.Marshal(reassignPayload)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	env.PRHandler.ReassignReviewer(w2, req)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var reassignResponse map[string]interface{}
	if err := json.NewDecoder(w2.Body).Decode(&reassignResponse); err != nil {
		t.Fatalf("failed to decode reassign response: %v", err)
	}

	newReviewerID := reassignResponse["replaced_by"].(string)
	updatedPR := reassignResponse["pr"].(map[string]interface{})
	newReviewers := updatedPR["assigned_reviewers"].([]interface{})

	if newReviewerID == "u30" {
		t.Errorf("new reviewer should not be the author")
	}

	if newReviewerID == oldReviewerID {
		t.Errorf("new reviewer should be different from old reviewer")
	}

	for _, reviewer := range newReviewers {
		if reviewer.(string) == oldReviewerID {
			t.Errorf("old reviewer %s should not be in the list after reassignment", oldReviewerID)
		}
	}

	found := false
	for _, reviewer := range newReviewers {
		if reviewer.(string) == newReviewerID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("new reviewer %s should be in the reviewers list", newReviewerID)
	}

	if len(newReviewers) != 2 {
		t.Errorf("expected 2 reviewers after reassignment, got %d", len(newReviewers))
	}
}

func TestReassignOnMergedPR(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "frontend", 4)

	w := CreateTestPR(t, env.PRHandler, "pr-3001", "Hotfix", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var createResponse map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResponse)
	pr := createResponse["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	oldReviewerID := reviewers[0].(string)

	mergePayload := `{"pull_request_id": "pr-3001"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString(mergePayload))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	env.PRHandler.MergePR(w2, req)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for merge, got %d", w2.Code)
	}

	reassignPayload := fmt.Sprintf(`{"pull_request_id": "pr-3001", "old_user_id": "%s"}`, oldReviewerID)
	req2 := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(reassignPayload))
	req2.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()

	env.PRHandler.ReassignReviewer(w3, req2)

	if w3.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict for merged PR, got %d", w3.Code)
	}

	var errorResponse map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&errorResponse)

	errorObj := errorResponse["error"].(map[string]interface{})
	if errorObj["code"] != "PR_MERGED" {
		t.Errorf("expected error code PR_MERGED, got %v", errorObj["code"])
	}
}

func TestReassignNonExistentReviewer(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "devops", 4)
	w := CreateTestPR(t, env.PRHandler, "pr-3002", "Feature", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	reassignPayload := `{"pull_request_id": "pr-3002", "old_user_id": "u99"}`
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(reassignPayload))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	env.PRHandler.ReassignReviewer(w2, req)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", w2.Code)
	}

	var errorResponse map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&errorResponse)

	errorObj := errorResponse["error"].(map[string]interface{})
	if errorObj["code"] != "NOT_ASSIGNED" {
		t.Errorf("expected error code NOT_ASSIGNED, got %v", errorObj["code"])
	}
}

func TestReassignNoCandidate(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "qa", 3)

	w := CreateTestPR(t, env.PRHandler, "pr-3003", "Test", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var createResponse map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResponse)
	pr := createResponse["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	if len(reviewers) == 0 {
		t.Skip("")
	}

	oldReviewerID := reviewers[0].(string)

	reassignPayload := fmt.Sprintf(`{"pull_request_id": "pr-3003", "old_user_id": "%s"}`, oldReviewerID)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString(reassignPayload))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	env.PRHandler.ReassignReviewer(w2, req)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", w2.Code)
	}

	var errorResponse map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&errorResponse)

	errorObj := errorResponse["error"].(map[string]interface{})
	if errorObj["code"] != "NO_CANDIDATE" {
		t.Errorf("expected error code NO_CANDIDATE, got %v", errorObj["code"])
	}
}
