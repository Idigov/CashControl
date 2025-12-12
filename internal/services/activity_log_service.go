package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"encoding/json"
	"errors"
	"log/slog"
)

type ActivityLogService interface {
	CreateActivityLog(req models.CreateActivityLogRequest) (*models.ActivityHistory, error)
	GetActivityLogs(filter models.ActivityFilter) ([]models.ActivityHistory, error)
}

type activityLogService struct {
	activityLog repository.ActivityLogRepository
	logger      *slog.Logger
}

func NewActivityLogService(activityLog repository.ActivityLogRepository, logger *slog.Logger) ActivityLogService {
	return &activityLogService{activityLog: activityLog, logger: logger}
}

func (s *activityLogService) CreateActivityLog(req models.CreateActivityLogRequest) (*models.ActivityHistory, error) {
	const op = "service.activity_log.create"

	if err := s.validateActivityLogCreate(req); err != nil {
		s.logger.Warn("validation failed for creating activity log",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(req.UserID)),
			slog.String("activity_type", string(req.ActivityType)),
			slog.String("entity_type", req.EntityType),
			slog.Uint64("entity_id", uint64(req.EntityID)),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		s.logger.Error("failed to marshal metadata",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(req.UserID)),
			slog.String("activity_type", string(req.ActivityType)),
			slog.String("entity_type", req.EntityType),
			slog.Uint64("entity_id", uint64(req.EntityID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	activityLog := &models.ActivityHistory{
		UserID:       req.UserID,
		ActivityType: req.ActivityType,
		EntityType:   req.EntityType,
		EntityID:     req.EntityID,
		Description:  req.Description,
		Metadata:     string(metadataJSON),
	}

	s.logger.Debug("creating activity log entry",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(activityLog.UserID)),
		slog.String("activity_type", string(activityLog.ActivityType)),
		slog.String("entity_type", activityLog.EntityType),
		slog.Uint64("entity_id", uint64(activityLog.EntityID)),
	)

	if err := s.activityLog.Create(activityLog); err != nil {
		s.logger.Error("failed to create activity log in repository",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(activityLog.UserID)),
			slog.String("activity_type", string(activityLog.ActivityType)),
			slog.String("entity_type", activityLog.EntityType),
			slog.Uint64("entity_id", uint64(activityLog.EntityID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("activity log created successfully",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(activityLog.UserID)),
		slog.String("activity_type", string(activityLog.ActivityType)),
		slog.String("entity_type", activityLog.EntityType),
		slog.Uint64("entity_id", uint64(activityLog.EntityID)),
	)

	return activityLog, nil
}

func (s *activityLogService) GetActivityLogs(filter models.ActivityFilter) ([]models.ActivityHistory, error) {
	const op = "service.activity_log.get"

	s.logger.Debug("retrieving activity logs",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(filter.UserID)),
	)

	activityLogs, err := s.activityLog.Get(filter)
	if err != nil {
		s.logger.Error("failed to retrieve activity logs",
			slog.String("op", op),
			slog.Uint64("user_id", uint64(filter.UserID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("retrieved activity logs successfully",
		slog.String("op", op),
		slog.Uint64("user_id", uint64(filter.UserID)),
		slog.Int("count", len(activityLogs)),
	)

	return activityLogs, nil
}

func (s *activityLogService) validateActivityLogCreate(req models.CreateActivityLogRequest) error {
	if req.UserID <= 0 {
		return errors.New("user_id must be greater than zero")
	}

	if req.ActivityType == "" {
		return errors.New("activity_type is required")
	}

	switch req.ActivityType {
	case models.ActivityTypeExpenseCreated,
		models.ActivityTypeExpenseUpdated,
		models.ActivityTypeExpenseDeleted,
		models.ActivityTypeCategoryCreated,
		models.ActivityTypeCategoryUpdated,
		models.ActivityTypeCategoryDeleted,
		models.ActivityTypeBudgetCreated,
		models.ActivityTypeBudgetUpdated,
		models.ActivityTypeRecurringCreated,
		models.ActivityTypeRecurringUpdated,
		models.ActivityTypeRecurringDeleted:
	default:
		return errors.New("invalid activity_type")
	}

	if req.EntityType == "" {
		return errors.New("entity_type is required")
	}

	if req.EntityID <= 0 {
		return errors.New("entity_id must be greater than zero")
	}

	if len(req.Description) > 255 {
		return errors.New("description is too long, max 255 characters")
	}

	return nil
}
