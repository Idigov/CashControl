package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

type AnalyticsHandler struct {
	service services.AnalyticsService
	logger  *slog.Logger
}

func NewAnalyticsHandler(service services.AnalyticsService, logger *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

func (h *AnalyticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/analytics", h.Get)
}

func (h *AnalyticsHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")

	period := models.AnalyticsPeriod(c.DefaultQuery("period", "day"))

	start, err := time.Parse("2006-01-02", c.Query("start"))
	if err != nil {
		h.logger.Warn("analytics invalid start date", slog.String("value", c.Query("start")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}

	end, err := time.Parse("2006-01-02", c.Query("end"))
	if err != nil {
		h.logger.Warn("analytics invalid end date", slog.String("value", c.Query("end")))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}

	h.logger.Info("analytics request",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("period", string(period)),
		slog.Time("start", start),
		slog.Time("end", end),
	)

	data, err := h.service.GetAnalytics(userID, period, start, end)
	if err != nil {
		h.logger.Warn("analytics failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("analytics success",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("points", len(data)),
	)

	c.JSON(http.StatusOK, data)
}
