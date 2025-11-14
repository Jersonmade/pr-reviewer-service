package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

type TeamService struct {
	storage *storage.PostgresStorage
}


func NewTeamService(s *storage.PostgresStorage) *TeamService {
	return &TeamService{
		storage: s,
	}
}


func (ts *TeamService) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	if team.TeamName == "" {
		return nil, errors.New("Team name cannot be empty")
	}

	if len(team.Members) == 0 {
    	return nil, errors.New("team must have at least one member")
	}

	for i, member := range team.Members {
		if member.UserID == "" {
			return nil, fmt.Errorf("member at index %d has empty user_id", i)
		}
		
		if member.Username == "" {
			return nil, fmt.Errorf("member at index %d has empty username", i)
		}
	}

	seen := make(map[string]bool)
	for _, member := range team.Members {
		if seen[member.UserID] {
			return nil, fmt.Errorf("duplicate user_id: %s", member.UserID)
		}
		seen[member.UserID] = true
	}

	err := ts.storage.CreateTeam(ctx, team)

	if err != nil {
		return nil, err
	}

	return ts.storage.GetTeam(ctx, team.TeamName)
}


func (ts *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	if teamName == "" {
		return nil, errors.New("Team name cannot be empty")
	}

	team, err := ts.storage.GetTeam(ctx, teamName)

	if err != nil {
		return nil, err
	}

	return team, nil
}


func (ts *TeamService) ValidateTeamExists(ctx context.Context, teamName string) error {
	_, err := ts.GetTeam(ctx, teamName)
	return err
}


func (ts *TeamService) GetTeamMemberCount(ctx context.Context, teamName string) (int, error) {
	team, err := ts.GetTeam(ctx, teamName)

	if err != nil {
		return 0, err
	}

	return len(team.Members), nil
}


func (ts *TeamService) GetActiveTeamMemberCount(ctx context.Context, teamName string) (int, error) {
	team, err := ts.GetTeam(ctx, teamName)
	
	if err != nil {
		return 0, err
	}

	count := 0
	for _, member := range team.Members {
		if member.IsActive {
			count++
		}
	}

	return count, nil
}


func (ts *TeamService) IsUserInTeam(ctx context.Context, teamName, userID string) (bool, error) {
	team, err := ts.GetTeam(ctx, teamName)
	
	if err != nil {
		return false, err
	}

	for _, member := range team.Members {
		if member.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}
