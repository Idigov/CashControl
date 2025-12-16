package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"time"
)

type StatisticsService interface {
	GetStatistics(
		userID uint,
		period models.StatisticsPeriod,
	) (*models.PeriodStatistics, error)
}

type statisticsService struct {
	repo repository.StatisticsRepository
}

func NewStatisticsService(repo repository.StatisticsRepository) StatisticsService {
	return &statisticsService{repo: repo}
}

func (s *statisticsService) GetStatistics(
	userID uint,
	period models.StatisticsPeriod,
) (*models.PeriodStatistics, error) {

	now := time.Now().UTC()
	var start time.Time

	switch period {
	case models.PeriodDay:
		start = time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0, time.UTC,
		)

	case models.PeriodWeek:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		base := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0, time.UTC,
		)
		start = base.AddDate(0, 0, -weekday+1)

	case models.PeriodMonth:
		start = time.Date(
			now.Year(), now.Month(), 1,
			0, 0, 0, 0, time.UTC,
		)

	case models.PeriodYear:
		start = time.Date(
			now.Year(), 1, 1,
			0, 0, 0, 0, time.UTC,
		)

	default:
		return nil, errors.New("invalid statistics period")
	}

	end := now
	stats, err := s.repo.GetPeriodStatistics(userID, start, end)
	if err != nil {
		return nil, err
	}
	
	// Устанавливаем период в результат
	stats.Period = period
	
	// Вычисляем среднее значение
	if stats.Count > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.Count)
	} else {
		stats.AverageAmount = 0
	}
	
	return stats, nil
}
