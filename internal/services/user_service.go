package services

import (
	"context"
	"errors"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

type UserService struct {
	storage *storage.PostgresStorage
}


func NewUserService(s *storage.PostgresStorage) *UserService {
	return &UserService{storage: s}
}


func (us *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, errors.New("User ID cannot be empty")
	}

	user, err := us.storage.GetUser(ctx, userID)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
            return nil, errors.New("USER_NOT_FOUND")
        }

		return nil, err
	}

	return user, nil
}


func (us *UserService) SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	if userID == "" {
		return nil, errors.New("User ID cannot be empty")
	}

	_, err := us.storage.GetUser(ctx, userID)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
            return nil, errors.New("USER_NOT_FOUND")
        }

		return nil, err
	}

	user, err := us.storage.UpdateUserActive(ctx, userID, isActive)

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
            return nil, errors.New("USER_NOT_FOUND")
        }
		
		return nil, err
	}

	return user, nil
}


func (us *UserService) GetActiveTeamMembers(ctx context.Context, teamName, excludeUserID string, excludeReviewers []string) ([]string, error) {
	if teamName == "" {
		return nil, errors.New("team_name cannot be empty")
	}

	return us.storage.GetActiveTeamMembers(ctx, teamName, excludeUserID, excludeReviewers)
}


func (us *UserService) ValidateUser(ctx context.Context, userID string) error {
	user, err := us.storage.GetUser(ctx, userID)

	if err != nil {
		return err
	}

	if !user.IsActive {
		return errors.New("User is not active")
	}

	return nil
}


func (us *UserService) GetUserTeamName(ctx context.Context, userID string) (string, error) {
	user, err := us.storage.GetUser(ctx, userID)

	if err != nil {
		return "", err
	}

	return user.TeamName, nil
}
