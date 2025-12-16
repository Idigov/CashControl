package repository

import (
	"cashcontrol/internal/models"
	"fmt"
	"log/slog"
	"os"
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
	db     *gorm.DB
	logger *slog.Logger
}

func NewStatisticsRepository(db *gorm.DB) StatisticsRepository {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return &gormStatisticsRepository{db: db, logger: logger}
}

func (r *gormStatisticsRepository) GetPeriodStatistics(
	userID uint,
	start, end time.Time,
) (*models.PeriodStatistics, error) {
	
	// Логируем параметры для отладки
	r.logger.Info("GetPeriodStatistics called",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("start", start.Format("2006-01-02 15:04:05")),
		slog.String("end", end.Format("2006-01-02 15:04:05")),
	)

	var rows []struct {
		CategoryID    uint
		CategoryName  string
		CategoryColor string
		TotalAmount   float64
		Count         int
	}

	// Используем BETWEEN как в analytics для совместимости с PostgreSQL
	query := `
		SELECT
			c.id    AS category_id,
			c.name  AS category_name,
			c.color AS category_color,
			COALESCE(SUM(e.amount), 0) AS total_amount,
			COUNT(e.id)   AS count
		FROM expenses e
		INNER JOIN categories c ON c.id = e.category_id
		WHERE e.user_id = ?
		  AND e.date BETWEEN ? AND ?
		  AND e.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.color
		HAVING COUNT(e.id) > 0
		ORDER BY total_amount DESC
	`
	
	r.logger.Info("Executing SQL query",
		slog.String("query", fmt.Sprintf("user_id=%d, start=%v, end=%v", userID, start, end)),
	)
	
	err := r.db.Raw(query, userID, start, end).Scan(&rows).Error
	
	if err != nil {
		r.logger.Error("SQL query failed",
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	r.logger.Info("SQL query succeeded",
		slog.Int("rows_count", len(rows)),
	)

	var total float64
	var count int

	for _, r := range rows {
		total += r.TotalAmount
		count += r.Count
	}

	// Вычисляем среднее значение
	averageAmount := 0.0
	if count > 0 {
		averageAmount = total / float64(count)
	}

	stats := &models.PeriodStatistics{
		StartDate:     start,
		EndDate:       end,
		TotalAmount:   total,
		Count:         count,
		AverageAmount: averageAmount,
		ByCategory:    []models.CategoryStatistics{}, // Инициализируем как пустой слайс, а не nil
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
