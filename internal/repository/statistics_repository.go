package repository

import (
	"cashcontrol/internal/models"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type StatisticsRepository interface {
	// GetPeriodStatistics получает статистику за период с группировкой по категориям
	GetPeriodStatistics(userID uint, startDate, endDate time.Time, categoryID *uint) ([]models.CategoryStatistics, float64, int, error)

	// GetCategoryStatistics получает статистику по категориям
	GetCategoryStatistics(userID uint, startDate, endDate *time.Time, categoryID *uint) ([]models.CategoryStatistics, error)

	// GetExpenseDistribution получает распределение расходов по категориям
	GetExpenseDistribution(userID uint, startDate, endDate *time.Time) ([]models.ExpenseDistribution, error)
}

type gormStatisticsRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewStatisticsRepository(db *gorm.DB, logger *slog.Logger) StatisticsRepository {
	return &gormStatisticsRepository{db: db, logger: logger}
}

// GetPeriodStatistics получает статистику за период с группировкой по категориям
func (r *gormStatisticsRepository) GetPeriodStatistics(userID uint, startDate, endDate time.Time, categoryID *uint) ([]models.CategoryStatistics, float64, int, error) {
	r.logger.Debug("repo.statistics.period",
		slog.String("op", "repo.statistics.period"),
		slog.Uint64("user_id", uint64(userID)),
		slog.Time("start_date", startDate),
		slog.Time("end_date", endDate),
	)

	var results []struct {
		CategoryID    uint    `gorm:"column:category_id"`
		CategoryName  string  `gorm:"column:category_name"`
		CategoryColor string  `gorm:"column:category_color"`
		TotalAmount   float64 `gorm:"column:total_amount"`
		Count         int     `gorm:"column:count"`
	}

	query := r.db.Model(&models.Expense{}).
		Select(`
			expenses.category_id,
			categories.name as category_name,
			categories.color as category_color,
			COALESCE(SUM(expenses.amount), 0) as total_amount,
			COUNT(expenses.id) as count
		`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ?", userID).
		Where("expenses.date >= ?", startDate).
		Where("expenses.date <= ?", endDate).
		Group("expenses.category_id, categories.name, categories.color")

	if categoryID != nil {
		query = query.Where("expenses.category_id = ?", *categoryID)
	}

	if err := query.Scan(&results).Error; err != nil {
		r.logger.Error("repo.statistics.period failed",
			slog.String("op", "repo.statistics.period"),
			slog.String("error", err.Error()),
		)
		return nil, 0, 0, err
	}

	// Вычисляем общую сумму и количество
	var totalAmount float64
	var totalCount int
	categoryStats := make([]models.CategoryStatistics, 0, len(results))

	for _, result := range results {
		totalAmount += result.TotalAmount
		totalCount += result.Count
		categoryStats = append(categoryStats, models.CategoryStatistics{
			CategoryID:    result.CategoryID,
			CategoryName:  result.CategoryName,
			CategoryColor: result.CategoryColor,
			TotalAmount:   result.TotalAmount,
			Count:         result.Count,
			Percentage:    0, // Будет вычислено в сервисе
		})
	}

	return categoryStats, totalAmount, totalCount, nil
}

// GetCategoryStatistics получает статистику по категориям
func (r *gormStatisticsRepository) GetCategoryStatistics(userID uint, startDate, endDate *time.Time, categoryID *uint) ([]models.CategoryStatistics, error) {
	r.logger.Debug("repo.statistics.categories",
		slog.String("op", "repo.statistics.categories"),
		slog.Uint64("user_id", uint64(userID)),
	)

	var results []struct {
		CategoryID    uint    `gorm:"column:category_id"`
		CategoryName  string  `gorm:"column:category_name"`
		CategoryColor string  `gorm:"column:category_color"`
		TotalAmount   float64 `gorm:"column:total_amount"`
		Count         int     `gorm:"column:count"`
	}

	query := r.db.Model(&models.Expense{}).
		Select(`
			expenses.category_id,
			categories.name as category_name,
			categories.color as category_color,
			COALESCE(SUM(expenses.amount), 0) as total_amount,
			COUNT(expenses.id) as count
		`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ?", userID).
		Group("expenses.category_id, categories.name, categories.color")

	if startDate != nil {
		query = query.Where("expenses.date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("expenses.date <= ?", *endDate)
	}
	if categoryID != nil {
		query = query.Where("expenses.category_id = ?", *categoryID)
	}

	if err := query.Scan(&results).Error; err != nil {
		r.logger.Error("repo.statistics.categories failed",
			slog.String("op", "repo.statistics.categories"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	categoryStats := make([]models.CategoryStatistics, 0, len(results))
	for _, result := range results {
		categoryStats = append(categoryStats, models.CategoryStatistics{
			CategoryID:    result.CategoryID,
			CategoryName:  result.CategoryName,
			CategoryColor: result.CategoryColor,
			TotalAmount:   result.TotalAmount,
			Count:         result.Count,
			Percentage:    0, // Будет вычислено в сервисе
		})
	}

	return categoryStats, nil
}

// GetExpenseDistribution получает распределение расходов по категориям
func (r *gormStatisticsRepository) GetExpenseDistribution(userID uint, startDate, endDate *time.Time) ([]models.ExpenseDistribution, error) {
	r.logger.Debug("repo.statistics.distribution",
		slog.String("op", "repo.statistics.distribution"),
		slog.Uint64("user_id", uint64(userID)),
	)

	var results []struct {
		CategoryID    uint    `gorm:"column:category_id"`
		CategoryName  string  `gorm:"column:category_name"`
		CategoryColor string  `gorm:"column:category_color"`
		Amount        float64 `gorm:"column:amount"`
	}

	query := r.db.Model(&models.Expense{}).
		Select(`
			expenses.category_id,
			categories.name as category_name,
			categories.color as category_color,
			COALESCE(SUM(expenses.amount), 0) as amount
		`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ?", userID).
		Group("expenses.category_id, categories.name, categories.color")

	if startDate != nil {
		query = query.Where("expenses.date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("expenses.date <= ?", *endDate)
	}

	if err := query.Scan(&results).Error; err != nil {
		r.logger.Error("repo.statistics.distribution failed",
			slog.String("op", "repo.statistics.distribution"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	distribution := make([]models.ExpenseDistribution, 0, len(results))
	for _, result := range results {
		distribution = append(distribution, models.ExpenseDistribution{
			CategoryID:    result.CategoryID,
			CategoryName:  result.CategoryName,
			CategoryColor: result.CategoryColor,
			Amount:        result.Amount,
			Percentage:    0, // Будет вычислено в сервисе
		})
	}

	return distribution, nil
}
