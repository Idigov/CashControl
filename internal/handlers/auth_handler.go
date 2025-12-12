package handlers

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/services"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service services.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(service services.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid register request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Register(req)
	if err != nil {
		h.logger.Warn("register failed", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid login request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(req)
	if err != nil {
		h.logger.Warn("login failed", slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
