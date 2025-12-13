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

type ExpenseHandler struct {
	service services.ExpenseService
	logger  *slog.Logger
}

func NewExpenseHandler(service services.ExpenseService, logger *slog.Logger) *ExpenseHandler {
	return &ExpenseHandler{service: service, logger: logger}
}

func (h *ExpenseHandler) RegisterRoutes(r *gin.Engine) {
	expenses := r.Group("/expenses")
	{
		expenses.GET("", h.List)
		expenses.POST("", h.Create)
		expenses.GET("/:id", h.Get)
		expenses.PATCH("/:id", h.Update)
		expenses.DELETE("/:id", h.Delete)
	}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	filter, _ := h.parseExpenseFilter(c) // Игнорируем ошибку парсинга фильтра, так как все поля опциональны

	expenses, err := h.service.GetExpenseList(filter)
	if err != nil {
		h.logger.Error("failed to get expense list",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense list retrieved",
		slog.Int("count", len(expenses)),
	)

	c.JSON(http.StatusOK, expenses)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
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

	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.CreateExpense(userID, req)
	if err != nil {
		h.logger.Warn("failed to create expense",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense created",
		slog.Uint64("expense_id", uint64(expense.ID)),
		slog.Uint64("user_id", uint64(userID)),
	)

	c.JSON(http.StatusCreated, expense)
}

func (h *ExpenseHandler) Get(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id))
	if err != nil {
		if err == services.ErrExpenseNotFound {
			h.logger.Warn("expense not found",
				slog.Uint64("expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to get expense",
			slog.Uint64("expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense retrieved",
		slog.Uint64("expense_id", id),
	)

	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.UpdateExpense(uint(id), req)
	if err != nil {
		h.logger.Warn("failed to update expense",
			slog.Uint64("expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense updated",
		slog.Uint64("expense_id", id),
	)

	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	if err := h.service.DeleteExpense(uint(id)); err != nil {
		h.logger.Error("failed to delete expense",
			slog.Uint64("expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("expense deleted",
		slog.Uint64("expense_id", id),
	)

	c.Status(http.StatusOK)
}

func (h *ExpenseHandler) parseExpenseFilter(c *gin.Context) (models.ExpenseFilter, error) {
	var filter models.ExpenseFilter

	// user_id обязателен
	if v := c.Query("user_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			filter.UserID = uint(id)
		}
	}

	if v := c.Query("category_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			categoryID := uint(id)
			filter.CategoryID = &categoryID
		}
	}
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.StartDate = &t
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.EndDate = &t
		}
	}
	if v := c.Query("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			filter.Limit = &l
		}
	}
	if v := c.Query("offset"); v != "" {
		if o, err := strconv.Atoi(v); err == nil {
			filter.Offset = &o
		}
	}

	return filter, nil
}
