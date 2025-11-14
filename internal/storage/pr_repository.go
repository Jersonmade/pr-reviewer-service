package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Jersonmade/pr-reviewer-service/internal/models"
)

func (s *PostgresStorage) CreatePR(ctx context.Context, pr *models.PullRequest) error {
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
		"SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)",
		pr.PullRequestID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return ErrPRExists
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, time.Now())

	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
            VALUES ($1, $2)
        `, pr.PullRequestID, reviewerID)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetPR(ctx context.Context, prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var createdAt time.Time
	var mergedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	pr.CreatedAt = &createdAt

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	rows, err := s.db.QueryContext(ctx, `
        SELECT reviewer_id
        FROM pr_reviewers
        WHERE pull_request_id = $1
        ORDER BY assigned_at
    `, prID)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	pr.AssignedReviewers = []string{}

	for rows.Next() {
		var reviewerID string

		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}

		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, nil
}

func (s *PostgresStorage) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	_, err := s.db.ExecContext(ctx, `
        UPDATE pull_requests
        SET status = 'MERGED', merged_at = CURRENT_TIMESTAMP
        WHERE pull_request_id = $1 AND status = 'OPEN'
    `, prID)

	if err != nil {
		return nil, err
	}

	return s.GetPR(ctx, prID)
}

func (s *PostgresStorage) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	tx, err := s.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("tx rollback failed: %v", err)
		}
	}()

	result, err := tx.ExecContext(ctx, `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1 AND reviewer_id = $2
    `, prID, oldReviewerID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotAssigned
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
        VALUES ($1, $2)
    `, prID, newReviewerID)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT DISTINCT p.pull_request_id, p.pull_request_name, p.author_id, p.status, p.created_at
        FROM pull_requests p
        INNER JOIN pr_reviewers pr ON p.pull_request_id = pr.pull_request_id
        WHERE pr.reviewer_id = $1
        ORDER BY p.created_at DESC
    `, userID)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	result := []models.PullRequestShort{}

	for rows.Next() {
		var pr models.PullRequestShort
		var createdAt time.Time
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, pr)
	}

	return result, nil
}
