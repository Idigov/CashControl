package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ActivityLogHandler struct {
	service services.ActivityLogService
	logger  *slog.Logger
}

func NewActivityLogHandler(service services.ActivityLogService, logger *slog.Logger) *ActivityLogHandler {
	return &ActivityLogHandler{service: service, logger: logger}
}

func (h *ActivityLogHandler) RegisterRoutes(r *gin.Engine) {
	logs := r.Group("/logs")
	{
		logs.GET("", h.Get)
		logs.POST("", h.Create)
	}
}

func (h *ActivityLogHandler) Get(c *gin.Context) {
	h.logger.Info("handling GET /logs",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_query", c.Request.URL.RawQuery),
	)

	filter, err := h.parseActivityFilter(c)
	if err != nil {
		h.logger.Warn("failed to parse filter",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("parsed filter",
		slog.Int("user_id", filter.UserID),
		slog.Any("activity_type", filter.ActivityType),
		slog.Any("entity_type", filter.EntityType),
		slog.Any("start_date", filter.StartDate),
		slog.Any("end_date", filter.EndDate),
		slog.Any("limit", filter.Limit),
		slog.Any("offset", filter.Offset),
	)

	logs, err := h.service.GetActivityLogs(filter)
	if err != nil {
		h.logger.Error("service.GetActivityLogs failed",
			slog.String("error", err.Error()),
			slog.Int("user_id", filter.UserID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("activity logs returned",
		slog.Int("count", len(logs)),
		slog.Int("user_id", filter.UserID),
	)

	c.JSON(http.StatusOK, logs)
}

func (h *ActivityLogHandler) Create(c *gin.Context) {
	h.logger.Info("handling POST /logs",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	var req models.CreateActivityLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid JSON body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	h.logger.Debug("request body parsed",
		slog.Int("user_id", req.UserID),
		slog.String("activity_type", string(req.ActivityType)),
		slog.String("entity_type", req.EntityType),
		slog.Int("entity_id", req.EntityID),
	)

	logEntry, err := h.service.CreateActivityLog(req)
	if err != nil {
		h.logger.Error("failed to create activity log",
			slog.String("error", err.Error()),
			slog.Int("user_id", req.UserID),
			slog.String("activity_type", string(req.ActivityType)),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create log entry"})
		return
	}

	h.logger.Info("activity log created",
		slog.Int("log_id", int(logEntry.ID)),
		slog.Int("user_id", logEntry.UserID),
		slog.String("activity_type", string(logEntry.ActivityType)),
	)

	c.JSON(http.StatusCreated, logEntry)
}

func (h *ActivityLogHandler) parseActivityFilter(c *gin.Context) (models.ActivityFilter, error) {
	var filter models.ActivityFilter

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		return filter, errors.New("user_id is required")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		return filter, errors.New("invalid user_id")
	}
	
	filter.UserID = userID

	if v := c.Query("activity_type"); v != "" {
		at := models.ActivityType(v)
		filter.ActivityType = &at
	}

	if v := c.Query("entity_type"); v != "" {
		filter.EntityType = &v
	}

	if v := c.Query("start_date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("invalid start_date")
		}
		filter.StartDate = &t
	}

	if v := c.Query("end_date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filter, errors.New("invalid end_date")
		}
		end := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &end
	}

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return filter, errors.New("invalid limit")
		}
		filter.Limit = &n
	}

	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return filter, errors.New("invalid offset")
		}
		filter.Offset = &n
	}

	return filter, nil
}
