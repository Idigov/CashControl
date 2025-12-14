package repository

import (
	"cashcontrol/internal/models"
	"time"

	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	GetAnalytics(
		userID uint,
		period models.AnalyticsPeriod,
		start, end time.Time,
	) ([]models.AnalyticsPoint, error)
}

type gormAnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &gormAnalyticsRepository{db: db}
}

func (r *gormAnalyticsRepository) GetAnalytics(
	userID uint,
	period models.AnalyticsPeriod,
	start, end time.Time,
) ([]models.AnalyticsPoint, error) {

	var trunc string

	switch period {
	case models.AnalyticsDay:
		trunc = "day"
	case models.AnalyticsWeek:
		trunc = "week"
	case models.AnalyticsMonth:
		trunc = "month"
	default:
		trunc = "day"
	}

	var result []models.AnalyticsPoint

	err := r.db.Raw(`
		SELECT
			date_trunc(?, date) AS date,
			SUM(amount) AS total,
			COUNT(*) AS count
		FROM expenses
		WHERE user_id = ?
		  AND date BETWEEN ? AND ?
		GROUP BY date
		ORDER BY date
	`, trunc, userID, start, end).Scan(&result).Error

	return result, err
}
