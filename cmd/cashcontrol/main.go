package main

import (
	"log/slog"
	"os"

	"cashcontrol/internal/config"
	"cashcontrol/internal/database"
	"cashcontrol/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := initLogger()

	if err := database.Init(cfg); err != nil {
		logger.Error("failed to init database", slog.String("error", err.Error()))
		panic(err)
	}

	if err := database.Migrate(); err != nil {
		logger.Error("failed to migrate database", slog.String("error", err.Error()))
		panic(err)
	}

	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("failed to close database", slog.String("error", err.Error()))
		}
	}()

	router := setupRouter(cfg, logger)

	router.Run(cfg.ServerAddress)
}

func setupRouter(cfg *config.Config, logger *slog.Logger) *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	handlers.RegisterRoutes(router, database.DB, logger, cfg)

	return router
}

func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
