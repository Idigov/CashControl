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

	// Инициализация маршрутизатора
	router := setupRouter(cfg, logger)

	router.Run(cfg.ServerAddress)
}

// setupRouter инициализирует все зависимости и настраивает роутер
func setupRouter(cfg *config.Config, logger *slog.Logger) *gin.Engine {
	// Настройка роутера
	router := gin.Default()

	handlers.RegisterRoutes(router, database.DB, logger, cfg)

	return router
}

// initLogger инициализирует структурированный логгер
func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
