package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RecurringExpenseHandler struct {
	service services.RecurringExpenseService
	logger  *slog.Logger
}

func NewRecurringExpenseHandler(service services.RecurringExpenseService, logger *slog.Logger) *RecurringExpenseHandler {
	return &RecurringExpenseHandler{service: service, logger: logger}
}

func (h *RecurringExpenseHandler) RegisterRoutes(r *gin.Engine) {
	recurringExpenses := r.Group("/recurring-expenses")
	{
		recurringExpenses.GET("", h.List)
		recurringExpenses.POST("", h.Create)
		recurringExpenses.GET("/active", h.GetActive)
		recurringExpenses.GET("/:id", h.Get)
		recurringExpenses.PATCH("/:id", h.Update)
		recurringExpenses.DELETE("/:id", h.Delete)
		recurringExpenses.POST("/:id/activate", h.Activate)
		recurringExpenses.POST("/:id/deactivate", h.Deactivate)
	}
}

func (h *RecurringExpenseHandler) List(c *gin.Context) {
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
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}

	recurringExpenses, err := h.service.GetRecurringExpenseList(userID)
	if err != nil {
		h.logger.Error("failed to get recurring expense list",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense list retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(recurringExpenses)),
	)

	c.JSON(http.StatusOK, recurringExpenses)
}

func (h *RecurringExpenseHandler) Create(c *gin.Context) {
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
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}

	var req models.CreateRecurringExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recurringExpense, err := h.service.CreateRecurringExpense(userID, req)
	if err != nil {
		h.logger.Warn("failed to create recurring expense",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense created",
		slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
		slog.Uint64("user_id", uint64(userID)),
		slog.String("type", string(recurringExpense.Type)),
	)

	c.JSON(http.StatusCreated, recurringExpense)
}

func (h *RecurringExpenseHandler) Get(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid recurring expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	recurringExpense, err := h.service.GetRecurringExpenseByID(uint(id))
	if err != nil {
		if err == services.ErrRecurringExpenseNotFound {
			h.logger.Warn("recurring expense not found",
				slog.Uint64("recurring_expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to get recurring expense",
			slog.Uint64("recurring_expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense retrieved",
		slog.Uint64("recurring_expense_id", id),
	)

	c.JSON(http.StatusOK, recurringExpense)
}

func (h *RecurringExpenseHandler) Update(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid recurring expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	var req models.UpdateRecurringExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recurringExpense, err := h.service.UpdateRecurringExpense(uint(id), req)
	if err != nil {
		if err == services.ErrRecurringExpenseNotFound {
			h.logger.Warn("recurring expense not found for update",
				slog.Uint64("recurring_expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Warn("failed to update recurring expense",
			slog.Uint64("recurring_expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense updated",
		slog.Uint64("recurring_expense_id", id),
	)

	c.JSON(http.StatusOK, recurringExpense)
}

func (h *RecurringExpenseHandler) Delete(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid recurring expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	if err := h.service.DeleteRecurringExpense(uint(id)); err != nil {
		if err == services.ErrRecurringExpenseNotFound {
			h.logger.Warn("recurring expense not found for delete",
				slog.Uint64("recurring_expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to delete recurring expense",
			slog.Uint64("recurring_expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense deleted",
		slog.Uint64("recurring_expense_id", id),
	)

	c.Status(http.StatusOK)
}

func (h *RecurringExpenseHandler) GetActive(c *gin.Context) {
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
	if err != nil {
		h.logger.Warn("invalid user_id parameter",
			slog.String("raw_user_id", userIDStr),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный user_id"})
		return
	}

	recurringExpenses, err := h.service.GetActiveRecurringExpenses(userID)
	if err != nil {
		h.logger.Error("failed to get active recurring expenses",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("active recurring expenses retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(recurringExpenses)),
	)

	c.JSON(http.StatusOK, recurringExpenses)
}

func (h *RecurringExpenseHandler) Activate(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid recurring expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	recurringExpense, err := h.service.ActivateRecurringExpense(uint(id))
	if err != nil {
		if err == services.ErrRecurringExpenseNotFound {
			h.logger.Warn("recurring expense not found for activation",
				slog.Uint64("recurring_expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to activate recurring expense",
			slog.Uint64("recurring_expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense activated",
		slog.Uint64("recurring_expense_id", id),
	)

	c.JSON(http.StatusOK, recurringExpense)
}

func (h *RecurringExpenseHandler) Deactivate(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid recurring expense id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	recurringExpense, err := h.service.DeactivateRecurringExpense(uint(id))
	if err != nil {
		if err == services.ErrRecurringExpenseNotFound {
			h.logger.Warn("recurring expense not found for deactivation",
				slog.Uint64("recurring_expense_id", id),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("failed to deactivate recurring expense",
			slog.Uint64("recurring_expense_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("recurring expense deactivated",
		slog.Uint64("recurring_expense_id", id),
	)

	c.JSON(http.StatusOK, recurringExpense)
}
