package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

func (s *PostgresStorage) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *PostgresStorage) UpdateUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`, isActive, userID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return s.GetUser(ctx, userID)
}

func (s *PostgresStorage) GetActiveTeamMembers(ctx context.Context, teamName, excludeUserID string, excludeReviewers []string) ([]string, error) {
	query := `
        SELECT user_id
        FROM users
        WHERE team_name = $1 AND is_active = true AND user_id != $2
    `
	args := []interface{}{teamName, excludeUserID}

	if len(excludeReviewers) > 0 {
		query += " AND user_id NOT IN ("
		for i, reviewerID := range excludeReviewers {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("$%d", i+3)
			args = append(args, reviewerID)
		}
		query += ")"
	}

	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	candidates := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		candidates = append(candidates, userID)
	}

	return candidates, nil
}
