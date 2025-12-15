package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service services.AnalyticsService
}

func NewAnalyticsHandler(service services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/analytics", h.Get)
}

func (h *AnalyticsHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")

	period := models.AnalyticsPeriod(c.DefaultQuery("period", "day"))

	start, err := time.Parse("2006-01-02", c.Query("start"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}

	end, err := time.Parse("2006-01-02", c.Query("end"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}

	data, err := h.service.GetAnalytics(userID, period, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
