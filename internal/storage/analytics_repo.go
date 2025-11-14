package storage

import (
	"context"
	"log"
)

func (s *PostgresStorage) GetReviewAssignmentsCount(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT reviewer_id, COUNT(*)
        FROM pr_reviewers
        GROUP BY reviewer_id
    `)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	counts := make(map[string]int)
	for rows.Next() {
		var reviewerID string
		var count int
		if err := rows.Scan(&reviewerID, &count); err != nil {
			return nil, err
		}
		counts[reviewerID] = count
	}
	return counts, nil
}
