package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BudgetHandler struct {
	service services.BudgetService
	logger  *slog.Logger
}

func NewBudgetHandler(service services.BudgetService, logger *slog.Logger) *BudgetHandler {
	return &BudgetHandler{service: service, logger: logger}
}

func (h *BudgetHandler) RegisterRoutes(r *gin.Engine) {
	budgets := r.Group("/budgets")
	{
		budgets.GET("", h.List)
		budgets.POST("", h.Create)
		budgets.GET("/status", h.GetStatus)
		budgets.GET("/by-month", h.GetByMonth)
		budgets.GET("/:id", h.Get)
		budgets.PATCH("/:id", h.Update)
		budgets.DELETE("/:id", h.Delete)
	}
}

func (h *BudgetHandler) List(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр user_id"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	budgets, err := h.service.GetBudgetList(userID)
	if err != nil {
		h.logger.Error("failed to get budget list",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget list retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(budgets)),
	)

	c.JSON(http.StatusOK, budgets)
}

func (h *BudgetHandler) Create(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр user_id"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	var req models.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.CreateBudget(userID, req)
	if err != nil {
		h.logger.Warn("failed to create budget",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget created",
		slog.Uint64("budget_id", uint64(budget.ID)),
		slog.Uint64("user_id", uint64(userID)),
		slog.Float64("amount", budget.Amount),
		slog.Int("month", budget.Month),
		slog.Int("year", budget.Year),
	)

	c.JSON(http.StatusCreated, budget)
}

func (h *BudgetHandler) Get(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid budget id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	budget, err := h.service.GetBudgetByID(uint(id))
	if err != nil {
		if err == services.ErrBudgetNotFound {
			h.logger.Warn("budget not found",
				slog.Uint64("budget_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to get budget",
			slog.Uint64("budget_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget retrieved",
		slog.Uint64("budget_id", id),
	)

	c.JSON(http.StatusOK, budget)
}

func (h *BudgetHandler) Update(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid budget id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	var req models.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.UpdateBudget(uint(id), req)
	if err != nil {
		if err == services.ErrBudgetNotFound {
			h.logger.Warn("budget not found for update",
				slog.Uint64("budget_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Warn("failed to update budget",
			slog.Uint64("budget_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget updated",
		slog.Uint64("budget_id", id),
	)

	c.JSON(http.StatusOK, budget)
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid budget id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	if err := h.service.DeleteBudget(uint(id)); err != nil {
		if err == services.ErrBudgetNotFound {
			h.logger.Warn("budget not found for delete",
				slog.Uint64("budget_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to delete budget",
			slog.Uint64("budget_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget deleted",
		slog.Uint64("budget_id", id),
	)

	c.Status(http.StatusOK)
}

func (h *BudgetHandler) GetStatus(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр user_id"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	monthStr := c.Query("month")
	if monthStr == "" {
		h.logger.Warn("missing month parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр month"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		h.logger.Warn("invalid month parameter",
			slog.String("raw_month", monthStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный month"})
		return
	}

	yearStr := c.Query("year")
	if yearStr == "" {
		h.logger.Warn("missing year parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр year"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		h.logger.Warn("invalid year parameter",
			slog.String("raw_year", yearStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный year"})
		return
	}

	status, err := h.service.GetBudgetStatus(userID, month, year)
	if err != nil {
		if err == services.ErrBudgetNotFound {
			h.logger.Warn("budget not found for status",
				slog.Uint64("user_id", uint64(userID)),
				slog.Int("month", month),
				slog.Int("year", year),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to get budget status",
			slog.Uint64("user_id", uint64(userID)),
			slog.Int("month", month),
			slog.Int("year", year),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget status retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("month", month),
		slog.Int("year", year),
		slog.Float64("spent", status.Spent),
		slog.Float64("remaining", status.Remaining),
		slog.Float64("percentage", status.Percentage),
	)

	c.JSON(http.StatusOK, status)
}

func (h *BudgetHandler) GetByMonth(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.logger.Warn("missing user_id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр user_id"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}
	userID := uint(userIDUint)

	monthStr := c.Query("month")
	if monthStr == "" {
		h.logger.Warn("missing month parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр month"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		h.logger.Warn("invalid month parameter",
			slog.String("raw_month", monthStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный month"})
		return
	}

	yearStr := c.Query("year")
	if yearStr == "" {
		h.logger.Warn("missing year parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "необходим параметр year"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		h.logger.Warn("invalid year parameter",
			slog.String("raw_year", yearStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный year"})
		return
	}

	budget, err := h.service.GetBudgetByUserIDAndMonth(userID, month, year)
	if err != nil {
		if err == services.ErrBudgetNotFound {
			h.logger.Warn("budget not found",
				slog.Uint64("user_id", uint64(userID)),
				slog.Int("month", month),
				slog.Int("year", year),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to get budget by month",
			slog.Uint64("user_id", uint64(userID)),
			slog.Int("month", month),
			slog.Int("year", year),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("budget retrieved by month",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("month", month),
		slog.Int("year", year),
		slog.Uint64("budget_id", uint64(budget.ID)),
	)

	c.JSON(http.StatusOK, budget)
}
