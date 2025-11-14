package services

import (
	"context"

	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

type AnalyticsService struct {
	storage *storage.PostgresStorage
}

func NewAnalyticsService(s *storage.PostgresStorage) *AnalyticsService {
	return &AnalyticsService{storage: s}
}

func (a *AnalyticsService) GetReviewAssignmentsCount(ctx context.Context) (map[string]int, error) {
	return a.storage.GetReviewAssignmentsCount(ctx)
}
