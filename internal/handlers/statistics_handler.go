package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type StatisticsHandler struct {
	service services.StatisticsService
	logger  *slog.Logger
}

func NewStatisticsHandler(service services.StatisticsService, logger *slog.Logger) *StatisticsHandler {
	return &StatisticsHandler{service: service, logger: logger}
}

func (h *StatisticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/statistics", h.Get)
}

func (h *StatisticsHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	if userID == 0 {
		h.logger.Warn("statistics user_id missing")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found in token",
		})
		return
	}

	period := models.StatisticsPeriod(
		c.DefaultQuery("period", string(models.PeriodMonth)),
	)

	h.logger.Info("statistics request",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
	)

	stats, err := h.service.GetStatistics(userID, period)
	if err != nil {
		h.logger.Warn("statistics failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("statistics success",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
	)

	c.JSON(http.StatusOK, stats)
}
