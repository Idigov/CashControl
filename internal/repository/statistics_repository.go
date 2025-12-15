package repository

import (
	"cashcontrol/internal/models"
	"time"

	"gorm.io/gorm"
)

type StatisticsRepository interface {
	GetPeriodStatistics(
		userID uint,
		start, end time.Time,
	) (*models.PeriodStatistics, error)
}

type gormStatisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) StatisticsRepository {
	return &gormStatisticsRepository{db: db}
}

func (r *gormStatisticsRepository) GetPeriodStatistics(
	userID uint,
	start, end time.Time,
) (*models.PeriodStatistics, error) {

	var rows []struct {
		CategoryID    uint
		CategoryName  string
		CategoryColor string
		TotalAmount   float64
		Count         int
	}

	err := r.db.Raw(`
		SELECT
			c.id    AS category_id,
			c.name  AS category_name,
			c.color AS category_color,
			SUM(e.amount) AS total_amount,
			COUNT(e.id)   AS count
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		WHERE e.user_id = ?
		  AND e.date BETWEEN ? AND ?
		GROUP BY c.id, c.name, c.color
		ORDER BY total_amount DESC
	`, userID, start, end).Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	var total float64
	var count int

	for _, r := range rows {
		total += r.TotalAmount
		count += r.Count
	}

	stats := &models.PeriodStatistics{
		StartDate:   start,
		EndDate:     end,
		TotalAmount: total,
		Count:       count,
	}

	for _, r := range rows {
		percentage := 0.0
		if total > 0 {
			percentage = (r.TotalAmount / total) * 100
		}

		stats.ByCategory = append(stats.ByCategory, models.CategoryStatistics{
			CategoryID:    r.CategoryID,
			CategoryName:  r.CategoryName,
			CategoryColor: r.CategoryColor,
			TotalAmount:   r.TotalAmount,
			Count:         r.Count,
			Percentage:    percentage,
		})
	}

	return stats, nil
}
