package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	service services.StatisticsService
}

func NewStatisticsHandler(service services.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{service: service}
}

func (h *StatisticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/statistics", h.Get)
}

func (h *StatisticsHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found in token",
		})
		return
	}

	period := models.StatisticsPeriod(
		c.DefaultQuery("period", string(models.PeriodMonth)),
	)

	stats, err := h.service.GetStatistics(userID, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
