package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"log/slog"
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
	logger *slog.Logger
}

func NewAnalyticsService(repo repository.AnalyticsRepository, logger *slog.Logger) AnalyticsService {
	return &analyticsService{repo: repo, logger: logger}
}

func (s *analyticsService) GetAnalytics(
	userID uint,
	period models.AnalyticsPeriod,
	start, end time.Time,
) ([]models.AnalyticsPoint, error) {

	if start.After(end) {
		if s.logger != nil {
			s.logger.Warn("analytics invalid range",
				slog.Uint64("user_id", uint64(userID)),
				slog.Time("start", start),
				slog.Time("end", end),
			)
		}
		return nil, errors.New("start date after end date")
	}

	data, err := s.repo.GetAnalytics(userID, period, start, end)
	if err != nil && s.logger != nil {
		s.logger.Error("analytics repo failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
	}
	return data, err
}
