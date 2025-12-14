package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	service services.StatisticsService
	logger  *slog.Logger
}

func NewStatisticsHandler(service services.StatisticsService, logger *slog.Logger) *StatisticsHandler {
	return &StatisticsHandler{service: service, logger: logger}
}

func (h *StatisticsHandler) RegisterRoutes(r *gin.Engine) {
	statistics := r.Group("/statistics")
	{
		statistics.GET("/period", h.GetPeriodStatistics)
		statistics.GET("/categories", h.GetCategoryStatistics)
		statistics.GET("/distribution", h.GetExpenseDistribution)
	}
}

// GetPeriodStatistics обрабатывает GET /statistics/period
// Параметры: user_id (обязательно), period (day/week/month/year), start_date, end_date, category_id
func (h *StatisticsHandler) GetPeriodStatistics(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	// Парсим user_id (обязательный параметр)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "параметр user_id обязателен"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id",
			slog.String("raw_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	// Парсим period (опционально, по умолчанию month)
	periodStr := c.DefaultQuery("period", string(models.PeriodMonth))
	period := models.StatisticsPeriod(periodStr)

	// Валидация периода
	if period != models.PeriodDay && period != models.PeriodWeek &&
		period != models.PeriodMonth && period != models.PeriodYear {
		h.logger.Warn("invalid period",
			slog.String("period", periodStr),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный период. Допустимые значения: day, week, month, year"})
		return
	}

	// Парсим даты (опционально)
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		} else {
			h.logger.Warn("invalid start_date format",
				slog.String("start_date", startDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат start_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		} else {
			h.logger.Warn("invalid end_date format",
				slog.String("end_date", endDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат end_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	// Парсим category_id (опционально)
	var categoryID *uint
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
			categoryIDUint := uint(id)
			categoryID = &categoryIDUint
		} else {
			h.logger.Warn("invalid category_id",
				slog.String("category_id", categoryIDStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный category_id"})
			return
		}
	}

	// Получаем статистику
	statistics, err := h.service.GetPeriodStatistics(userID, period, startDate, endDate, categoryID)
	if err != nil {
		h.logger.Error("failed to get period statistics",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("period statistics retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
	)

	c.JSON(http.StatusOK, statistics)
}

// GetCategoryStatistics обрабатывает GET /statistics/categories
// Параметры: user_id (обязательно), start_date, end_date, category_id
func (h *StatisticsHandler) GetCategoryStatistics(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	// Парсим user_id (обязательный параметр)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "параметр user_id обязателен"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id",
			slog.String("raw_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	// Парсим даты (опционально)
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		} else {
			h.logger.Warn("invalid start_date format",
				slog.String("start_date", startDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат start_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		} else {
			h.logger.Warn("invalid end_date format",
				slog.String("end_date", endDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат end_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	// Парсим category_id (опционально)
	var categoryID *uint
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
			categoryIDUint := uint(id)
			categoryID = &categoryIDUint
		} else {
			h.logger.Warn("invalid category_id",
				slog.String("category_id", categoryIDStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный category_id"})
			return
		}
	}

	// Получаем статистику
	statistics, err := h.service.GetCategoryStatistics(userID, startDate, endDate, categoryID)
	if err != nil {
		h.logger.Error("failed to get category statistics",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category statistics retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("categories_count", len(statistics)),
	)

	c.JSON(http.StatusOK, statistics)
}

// GetExpenseDistribution обрабатывает GET /statistics/distribution
// Параметры: user_id (обязательно), start_date, end_date
func (h *StatisticsHandler) GetExpenseDistribution(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	// Парсим user_id (обязательный параметр)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "параметр user_id обязателен"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id",
			slog.String("raw_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	// Парсим даты (опционально)
	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		} else {
			h.logger.Warn("invalid start_date format",
				slog.String("start_date", startDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат start_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &t
		} else {
			h.logger.Warn("invalid end_date format",
				slog.String("end_date", endDateStr),
				slog.String("reason", err.Error()),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат end_date. Используйте YYYY-MM-DD"})
			return
		}
	}

	// Получаем распределение
	distribution, err := h.service.GetExpenseDistribution(userID, startDate, endDate)
	if err != nil {
		h.logger.Error("failed to get expense distribution",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense distribution retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("categories_count", len(distribution)),
	)

	c.JSON(http.StatusOK, distribution)
}
