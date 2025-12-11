package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var errActivityLogNil = errors.New("activity log is nil")

type ActivityLogRepository interface {
	Get(filter *models.ActivityFilter) ([]models.ActivityHistory, error)
	Create(logEntry *models.ActivityHistory) error
}

type activityLogRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewActivityLogRepository(db *gorm.DB, logger *slog.Logger) ActivityLogRepository {
	return &activityLogRepository{db: db, logger: logger}
}

// Create сохраняет запись об активности
func (r *activityLogRepository) Create(logEntry *models.ActivityHistory) error {
	const op = "repo.activity_log.create"

	if logEntry == nil {
		r.logger.Warn("activity log is nil", slog.String("op", op))
		return errActivityLogNil
	}

	// Debug: перед созданием
	r.logger.Debug("creating activity log",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(logEntry.UserID)),
		slog.String("activity_type", string(logEntry.ActivityType)),
		slog.String("entity_type", logEntry.EntityType),
		slog.Uint64("entity_id", uint64(logEntry.EntityID)),
	)

	// Создание записи
	if err := r.db.Create(&logEntry).Error; err != nil {
		r.logger.Error("failed to create activity log",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(logEntry.UserID)),
			slog.String("activity_type", string(logEntry.ActivityType)),
			slog.String("entity_type", logEntry.EntityType),
			slog.Uint64("entity_id", uint64(logEntry.EntityID)),
			slog.String("error", err.Error()),
		)
		return err
	}

	// Info: успешное создание
	r.logger.Info("activity log created",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(logEntry.UserID)),
		slog.String("activity_type", string(logEntry.ActivityType)),
		slog.String("entity_type", logEntry.EntityType),
		slog.Uint64("entity_id", uint64(logEntry.EntityID)),
	)

	return nil
}

// Get возвращает список записей по фильтру
func (r *activityLogRepository) Get(filter *models.ActivityFilter) ([]models.ActivityHistory, error) {
	const op = "repo.activity_log.get"

	r.logger.Debug("retrieving activity logs",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(filter.UserID)),
		slog.String("activity_type", func() string {
			if filter.ActivityType != nil {
				return string(*filter.ActivityType)
			}
			return "all"
		}()),
		slog.String("entity_type", func() string {
			if filter.EntityType != nil {
				return *filter.EntityType
			}
			return "all"
		}()),
		slog.String("start_date", func() string {
			if filter.StartDate != nil {
				return filter.StartDate.Format("2006-01-02")
			}
			return ""
		}()),
		slog.String("end_date", func() string {
			if filter.EndDate != nil {
				return filter.EndDate.Format("2006-01-02")
			}
			return ""
		}()),
	)

	var logs []models.ActivityHistory
	query := r.db.Model(&models.ActivityHistory{}).Where("user_id = ?", filter.UserID)

	if filter.ActivityType != nil {
		query = query.Where("activity_type = ?", *filter.ActivityType)
	}
	if filter.EntityType != nil {
		query = query.Where("entity_type = ?", *filter.EntityType)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}

	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		r.logger.Error("failed to retrieve activity logs",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(filter.UserID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	r.logger.Debug("retrieved activity logs",
		slog.String("op", op),
		slog.Int("count", len(logs)),
	)

	return logs, nil
}
