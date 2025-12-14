package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"time"
)

type AnalyticsService interface {
	GetAnalytics(
		userID uint,
		period models.AnalyticsPeriod,
		start, end time.Time,
	) ([]models.AnalyticsPoint, error)
}

type analyticsService struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) GetAnalytics(
	userID uint,
	period models.AnalyticsPeriod,
	start, end time.Time,
) ([]models.AnalyticsPoint, error) {

	if start.After(end) {
		return nil, errors.New("start date after end date")
	}

	return s.repo.GetAnalytics(userID, period, start, end)
}
