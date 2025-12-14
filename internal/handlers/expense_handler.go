package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"fmt"
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

func (h *ExpenseHandler) RegisterRoutes(r *gin.RouterGroup) {
	expenses := r.Group("/expenses")
	{
		expenses.GET("", h.List)
		expenses.POST("", h.Create)
		expenses.GET("/:id", h.Get)
		expenses.PATCH("/:id", h.Update)
		expenses.DELETE("/:id", h.Delete)

	}
}

// -------- LIST --------

func (h *ExpenseHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	filter, _ := h.parseExpenseFilter(c)
	filter.UserID = userID

	expenses, err := h.service.GetExpenseList(filter)
	if err != nil {
		h.logger.Error("failed to list expenses", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

// -------- CREATE --------

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.CreateExpense(userID, req)
	fmt.Println("categoryID",expense.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

// -------- GET --------

func (h *ExpenseHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id))
	if err != nil {
		if err == services.ErrExpenseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if expense.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

// -------- UPDATE --------

func (h *ExpenseHandler) Update(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id))
	if err != nil {
		if err == services.ErrExpenseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if expense.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateExpense(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// -------- DELETE --------

func (h *ExpenseHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id))
	if err != nil {
		if err == services.ErrExpenseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if expense.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.DeleteExpense(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// -------- FILTER --------

func (h *ExpenseHandler) parseExpenseFilter(c *gin.Context) (models.ExpenseFilter, error) {
	var filter models.ExpenseFilter

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
