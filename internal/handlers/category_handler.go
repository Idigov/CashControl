package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service services.CategoryService
	logger  *slog.Logger
}

func NewCategoryHandler(service services.CategoryService, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{service: service, logger: logger}
}

func (h *CategoryHandler) RegisterRoutes(r *gin.RouterGroup) {
	categories := r.Group("/categories")
	{
		categories.GET("", h.List)
		categories.POST("", h.Create)
		categories.GET("/:id", h.Get)
		categories.PATCH("/:id", h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}

func (h *CategoryHandler) List(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userID := c.GetUint("user_id")

	categories, err := h.service.GetCategoryList(userID)
	if err != nil {
		h.logger.Error("failed to get category list",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category list retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(categories)),
	)

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userID := c.GetUint("user_id")

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.CreateCategory(userID, req)
	if err != nil {
		h.logger.Warn("failed to create category",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category created",
		slog.Uint64("category_id", uint64(category.ID)),
		slog.Uint64("user_id", uint64(userID)),
	)

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Get(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := h.service.GetCategoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	fmt.Println("cat_handler---------------", category)
	c.JSON(http.StatusOK, category)
}


func (h *CategoryHandler) Update(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := h.service.GetCategoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateCategory(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}


func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := h.service.GetCategoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if category.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

