package services

import (
	"context"

	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

type StatsService struct {
	storage *storage.PostgresStorage
}

func NewStatsService(s *storage.PostgresStorage) *StatsService {
	return &StatsService{storage: s}
}

func (a *StatsService) GetReviewAssignmentsCount(ctx context.Context) (map[string]int, error) {
	return a.storage.GetReviewAssignmentsCount(ctx)
}
