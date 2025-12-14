package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"log/slog"
	"time"
)

type StatisticsService interface {
	// GetPeriodStatistics получает статистику за период (день/неделя/месяц/год)
	GetPeriodStatistics(userID uint, period models.StatisticsPeriod, startDate, endDate *time.Time, categoryID *uint) (*models.PeriodStatistics, error)

	// GetCategoryStatistics получает статистику по категориям
	GetCategoryStatistics(userID uint, startDate, endDate *time.Time, categoryID *uint) ([]models.CategoryStatistics, error)

	// GetExpenseDistribution получает распределение расходов по категориям
	GetExpenseDistribution(userID uint, startDate, endDate *time.Time) ([]models.ExpenseDistribution, error)
}

type statisticsService struct {
	statsRepo repository.StatisticsRepository
	logger    *slog.Logger
}

func NewStatisticsService(statsRepo repository.StatisticsRepository, logger *slog.Logger) StatisticsService {
	return &statisticsService{statsRepo: statsRepo, logger: logger}
}

// GetPeriodStatistics получает статистику за период
func (s *statisticsService) GetPeriodStatistics(userID uint, period models.StatisticsPeriod, startDate, endDate *time.Time, categoryID *uint) (*models.PeriodStatistics, error) {
	s.logger.Debug("statistics.period",
		slog.String("op", "statistics.period"),
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
	)

	// Определяем даты периода, если не указаны
	var actualStartDate, actualEndDate time.Time
	now := time.Now()

	switch period {
	case models.PeriodDay:
		if startDate != nil {
			actualStartDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		} else {
			actualStartDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		}
		actualEndDate = actualStartDate.Add(24 * time.Hour).Add(-time.Second)

	case models.PeriodWeek:
		if startDate != nil {
			actualStartDate = *startDate
		} else {
			// Начало текущей недели (понедельник)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // Воскресенье = 7
			}
			actualStartDate = now.AddDate(0, 0, -weekday+1)
			actualStartDate = time.Date(actualStartDate.Year(), actualStartDate.Month(), actualStartDate.Day(), 0, 0, 0, 0, actualStartDate.Location())
		}
		actualEndDate = actualStartDate.AddDate(0, 0, 7).Add(-time.Second)

	case models.PeriodMonth:
		if startDate != nil {
			actualStartDate = *startDate
		} else {
			actualStartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
		actualEndDate = actualStartDate.AddDate(0, 1, 0).Add(-time.Second)

	case models.PeriodYear:
		if startDate != nil {
			actualStartDate = *startDate
		} else {
			actualStartDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		}
		actualEndDate = actualStartDate.AddDate(1, 0, 0).Add(-time.Second)

	default:
		// Если период не указан, используем переданные даты или текущий месяц
		if startDate != nil && endDate != nil {
			actualStartDate = *startDate
			actualEndDate = *endDate
		} else {
			actualStartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			actualEndDate = actualStartDate.AddDate(0, 1, 0).Add(-time.Second)
		}
	}

	// Если endDate передан явно, используем его
	if endDate != nil {
		actualEndDate = *endDate
	}

	// Получаем статистику из репозитория
	categoryStats, totalAmount, totalCount, err := s.statsRepo.GetPeriodStatistics(userID, actualStartDate, actualEndDate, categoryID)
	if err != nil {
		s.logger.Error("failed to get period statistics",
			slog.String("op", "statistics.period"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Вычисляем проценты для каждой категории
	for i := range categoryStats {
		if totalAmount > 0 {
			categoryStats[i].Percentage = (categoryStats[i].TotalAmount / totalAmount) * 100
		}
	}

	// Вычисляем среднюю сумму
	var averageAmount float64
	if totalCount > 0 {
		averageAmount = totalAmount / float64(totalCount)
	}

	statistics := &models.PeriodStatistics{
		Period:        period,
		StartDate:     actualStartDate,
		EndDate:       actualEndDate,
		TotalAmount:   totalAmount,
		Count:         totalCount,
		AverageAmount: averageAmount,
		ByCategory:    categoryStats,
	}

	s.logger.Info("period statistics retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
		slog.Float64("total_amount", totalAmount),
		slog.Int("count", totalCount),
	)

	return statistics, nil
}

// GetCategoryStatistics получает статистику по категориям
func (s *statisticsService) GetCategoryStatistics(userID uint, startDate, endDate *time.Time, categoryID *uint) ([]models.CategoryStatistics, error) {
	s.logger.Debug("statistics.categories",
		slog.String("op", "statistics.categories"),
		slog.Uint64("user_id", uint64(userID)),
	)

	categoryStats, err := s.statsRepo.GetCategoryStatistics(userID, startDate, endDate, categoryID)
	if err != nil {
		s.logger.Error("failed to get category statistics",
			slog.String("op", "statistics.categories"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Вычисляем общую сумму для расчета процентов
	var totalAmount float64
	for _, stat := range categoryStats {
		totalAmount += stat.TotalAmount
	}

	// Вычисляем проценты для каждой категории
	for i := range categoryStats {
		if totalAmount > 0 {
			categoryStats[i].Percentage = (categoryStats[i].TotalAmount / totalAmount) * 100
		}
	}

	s.logger.Info("category statistics retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("categories_count", len(categoryStats)),
	)

	return categoryStats, nil
}

// GetExpenseDistribution получает распределение расходов по категориям
func (s *statisticsService) GetExpenseDistribution(userID uint, startDate, endDate *time.Time) ([]models.ExpenseDistribution, error) {
	s.logger.Debug("statistics.distribution",
		slog.String("op", "statistics.distribution"),
		slog.Uint64("user_id", uint64(userID)),
	)

	distribution, err := s.statsRepo.GetExpenseDistribution(userID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get expense distribution",
			slog.String("op", "statistics.distribution"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Вычисляем общую сумму для расчета процентов
	var totalAmount float64
	for _, dist := range distribution {
		totalAmount += dist.Amount
	}

	// Вычисляем проценты для каждой категории
	for i := range distribution {
		if totalAmount > 0 {
			distribution[i].Percentage = (distribution[i].Amount / totalAmount) * 100
		}
	}

	s.logger.Info("expense distribution retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("categories_count", len(distribution)),
	)

	return distribution, nil
}
