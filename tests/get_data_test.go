package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPR(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "backend", 5)
	w := CreateTestPR(t, env.PRHandler, "pr-4000", "Add logging", "u30")

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create PR: %d", w.Code)
	}

	pr, err := env.Store.GetPR(context.Background(), "pr-4000")
	if err != nil {
		t.Fatalf("failed to get PR: %v", err)
	}

	if pr.PullRequestID != "pr-4000" {
		t.Errorf("expected pr-4000, got %s", pr.PullRequestID)
	}

	if pr.PullRequestName != "Add logging" {
		t.Errorf("expected 'Add logging', got %s", pr.PullRequestName)
	}

	if pr.AuthorID != "u30" {
		t.Errorf("expected author u30, got %s", pr.AuthorID)
	}

	if pr.Status != "OPEN" {
		t.Errorf("expected status OPEN, got %s", pr.Status)
	}

	if len(pr.AssignedReviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
	}

	if pr.CreatedAt == nil {
		t.Error("createdAt should not be nil")
	}
}

func TestGetNonExistentPR(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	_, err := env.Store.GetPR(context.Background(), "pr-99999")

	if err == nil {
		t.Error("expected error for non-existent PR, got nil")
	}
}

func TestGetTeam(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "qa", 4)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=qa", nil)
	w := httptest.NewRecorder()

	env.TeamHandler.GetTeam(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var team map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&team); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if team["team_name"] != "qa" {
		t.Errorf("expected team name 'qa', got %v", team["team_name"])
	}

	members := team["members"].([]interface{})

	if len(members) != 4 {
		t.Fatalf("expected 4 members, got %d", len(members))
	}

	for i, member := range members {
		m := member.(map[string]interface{})
		expectedUserID := fmt.Sprintf("u%d", 30+i)

		if m["user_id"] != expectedUserID {
			t.Errorf("member %d: expected user_id %s, got %v", i, expectedUserID, m["user_id"])
		}

		if m["is_active"] != true {
			t.Errorf("member %d should be active", i)
		}
	}
}

func TestGetNonExistentTeam(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()

	env.TeamHandler.GetTeam(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var errorResponse map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&errorResponse)

	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	errorObj := errorResponse["error"].(map[string]interface{})
	if errorObj["code"] != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %v", errorObj["code"])
	}
}

func TestGetPRsByReviewer(t *testing.T) {
	env := SetupTestEnvironment(t)
	defer env.Cleanup()

	CreateTestTeam(t, env.TeamHandler, "frontend", 5)

	w1 := CreateTestPR(t, env.PRHandler, "pr-5000", "Feature A", "u30")

	var createResponse map[string]interface{}
	err := json.NewDecoder(w1.Body).Decode(&createResponse)

	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	pr := createResponse["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	if len(reviewers) == 0 {
		t.Skip("No reviewers assigned, skipping test")
	}

	reviewerID := reviewers[0].(string)

	CreateTestPR(t, env.PRHandler, "pr-5001", "Feature B", "u31")
	CreateTestPR(t, env.PRHandler, "pr-5002", "Feature C", "u32")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/getReview?user_id=%s", reviewerID), nil)
	w := httptest.NewRecorder()

	env.UserHandler.GetUserReviews(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	prs := response["pull_requests"].([]interface{})

	if len(prs) == 0 {
		t.Errorf("expected at least 1 PR for reviewer %s", reviewerID)
	}

	for i, pr := range prs {
		p := pr.(map[string]interface{})

		if p["pull_request_id"] == nil {
			t.Errorf("PR %d: pull_request_id is missing", i)
		}

		if p["author_id"] == nil {
			t.Errorf("PR %d: author_id is missing", i)
		}

		if p["status"] == nil {
			t.Errorf("PR %d: status is missing", i)
		}
	}
}
