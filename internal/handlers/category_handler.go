package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
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

func (h *CategoryHandler) RegisterRoutes(r *gin.Engine) {
	categories := r.Group("/categories")
	{
		categories.GET("/:userId", h.ListByUser)
		categories.POST("/:userId", h.Create)
		categories.GET("/detail/:id", h.Get)
		categories.PATCH("/:id", h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}

func (h *CategoryHandler) ListByUser(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid user id",
			slog.String("raw_id", c.Param("userId")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор пользователя"})
		return
	}

	categories, err := h.service.GetCategoryList(int(userID))
	if err != nil {
		h.logger.Error("failed to get category list",
			slog.Int("user_id", int(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category list retrieved",
		slog.Int("user_id", int(userID)),
		slog.Int("count", len(categories)),
	)

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid user id",
			slog.String("raw_id", c.Param("userId")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор пользователя"})
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.CreateCategory(int(userID), req)
	if err != nil {
		h.logger.Warn("failed to create category",
			slog.Int("user_id", int(userID)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category created",
		slog.Uint64("category_id", uint64(category.ID)),
		slog.Int("user_id", int(userID)),
	)

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Get(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid category id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	category, err := h.service.GetCategoryByID(uint(id))
	if err != nil {
		h.logger.Error("failed to get category",
			slog.Uint64("category_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category retrieved",
		slog.Uint64("category_id", id),
	)

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid category id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.service.UpdateCategory(uint(id), req)
	if err != nil {
		h.logger.Warn("failed to update category",
			slog.Uint64("category_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category updated",
		slog.Uint64("category_id", id),
	)

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid category id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	err = h.service.DeleteCategory(uint(id))
	if err != nil {
		h.logger.Warn("failed to delete category",
			slog.Uint64("category_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("category deleted",
		slog.Uint64("category_id", id),
	)

	c.JSON(http.StatusOK, gin.H{"message": "категория удалена"})
}
