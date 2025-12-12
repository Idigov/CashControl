package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service services.UserService
	logger  *slog.Logger
}

func NewUserHandler(service services.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.Get)
		users.PATCH("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}

func (h *UserHandler) List(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	users, err := h.service.GetUserList()
	if err != nil {
		h.logger.Error("failed to get user list",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user list retrieved",
		slog.Int("count", len(users)),
	)

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
	)

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(req)
	if err != nil {
		h.logger.Warn("failed to create user",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user created",
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("email", user.Email),
	)

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Get(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid user id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		h.logger.Error("failed to get user",
			slog.Uint64("user_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user retrieved",
		slog.Uint64("user_id", id),
	)

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid user id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	var req struct {
		Email    string `json:"email,omitempty"`
		Username string `json:"username,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request body",
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(uint(id), req.Email, req.Username)
	if err != nil {
		h.logger.Warn("failed to update user",
			slog.Uint64("user_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user updated",
		slog.Uint64("user_id", id),
	)

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	h.logger.Info("incoming request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("raw_id", c.Param("id")),
	)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid user id",
			slog.String("raw_id", c.Param("id")),
			slog.String("reason", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный идентификатор"})
		return
	}

	err = h.service.DeleteUser(uint(id))
	if err != nil {
		h.logger.Warn("failed to delete user",
			slog.Uint64("user_id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user deleted",
		slog.Uint64("user_id", id),
	)

	c.JSON(http.StatusOK, gin.H{"message": "пользователь удален"})
}
