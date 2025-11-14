package services

import (
	"context"
	"errors"
	"math/rand"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

type PRService struct {
	storage *storage.PostgresStorage
	userService *UserService
}

func (prs *PRService) NewPRService(s *storage.PostgresStorage, us *UserService) (*PRService) {
	return &PRService{
		storage: s,
		userService: us,
	}
}

func (ps *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, errors.New("pull_request_id cannot be empty")
	}

	if prName == "" {
		return nil, errors.New("pull_request_name cannot be empty")
	}

	if authorID == "" {
		return nil, errors.New("author_id cannot be empty")
	}

	author, err := ps.userService.GetUser(ctx, authorID)

	if err != nil {
    	return nil, err
	}

	reviewers, err := ps.selectReviewers(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
	}

	if err := ps.storage.CreatePR(ctx, pr); err != nil {
		return nil, err
	}

	return ps.storage.GetPR(ctx, prID)
}


func (ps *PRService) selectReviewers(ctx context.Context, teamName, excludeUserID string) ([]string, error) {
	candidates, err := ps.userService.GetActiveTeamMembers(ctx, teamName, excludeUserID, []string{})
	
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
    	return []string{}, nil
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	max := 2
	if len(candidates) < max {
		max = len(candidates)
	}

	return candidates[:max], nil
}


func (ps *PRService) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, errors.New("pull_request_id cannot be empty")
	}

	pr, err := ps.storage.GetPR(ctx, prID)

	if err != nil {
		return nil, err
	}

	return pr, nil
}


func (ps *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, errors.New("pull_request_id cannot be empty")
	}

	pr, err := ps.storage.GetPR(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == "MERGED" {
		return pr, nil
	}

	return ps.storage.MergePR(ctx, prID)
}


func (ps *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (string, error) {
	if prID == "" {
		return "", errors.New("pull_request_id cannot be empty")
	}

	if oldReviewerID == "" {
		return "", errors.New("old_user_id cannot be empty")
	}

	pr, err := ps.storage.GetPR(ctx, prID)

	if err != nil {
		return "", err
	}

	if pr.Status == "MERGED" {
		return "", errors.New("PR_MERGED")
	}

	found := false
	for _, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("NOT_ASSIGNED")
	}

	oldReviewer, err := ps.userService.GetUser(ctx, oldReviewerID)

	if err != nil {
		return "", err
	}

	candidates, err := ps.userService.GetActiveTeamMembers(ctx, oldReviewer.TeamName, pr.AuthorID, pr.AssignedReviewers)
	
	if err != nil {
		return "", err
	}

	if len(candidates) == 0 {
		return "", errors.New("NO_CANDIDATE")
	}

	newReviewerID := candidates[rand.Intn(len(candidates))]

	if err := ps.storage.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		if err.Error() == "NOT_ASSIGNED" {
			return "", errors.New("NOT_ASSIGNED")
		}
		return "", err
	}

	return newReviewerID, nil
}


func (ps *PRService) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	if userID == "" {
		return nil, errors.New("user_id cannot be empty")
	}

	if _, err := ps.userService.GetUser(ctx, userID); err != nil {
		return nil, err
	}

	return ps.storage.GetPRsByReviewer(ctx, userID)
}


func (ps *PRService) ValidatePRExists(ctx context.Context, prID string) error {
	_, err := ps.GetPR(ctx, prID)
	return err
}


func (ps *PRService) IsPRMerged(ctx context.Context, prID string) (bool, error) {
	pr, err := ps.GetPR(ctx, prID)
	
	if err != nil {
		return false, err
	}

	return pr.Status == "MERGED", nil
}