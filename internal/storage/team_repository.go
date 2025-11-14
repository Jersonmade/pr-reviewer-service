package storage

import (
	"context"
	"database/sql"
	"log"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

func (s *PostgresStorage) CreateTeam(ctx context.Context, team *models.Team) error {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("tx rollback failed: %v", err)
		}
	}()

	var exists bool

	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		team.TeamName,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return ErrTeamExists
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO teams (team_name) VALUES ($1)",
		team.TeamName,
	)

	if err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active,
				updated_at = CURRENT_TIMESTAMP
			`, member.UserID, member.Username, team.TeamName, member.IsActive)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	var exists bool

	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		teamName,
	).Scan(&exists)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrNotFound
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`, teamName)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	members := []models.TeamMember{}
	for rows.Next() {
		var m models.TeamMember

		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}

		members = append(members, m)
	}

	return &models.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
