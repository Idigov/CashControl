package main

import (
	"log/slog"
	"os"

	"cashcontrol/internal/config"
	"cashcontrol/internal/database"
	"cashcontrol/internal/handlers"
	"cashcontrol/internal/repository"
	"cashcontrol/internal/services"
	"log/slog"
	"os"


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
	router := setupRouter()

	router.Run(cfg.ServerAddress)
}

// setupRouter инициализирует все зависимости и настраивает роутер
func setupRouter() *gin.Engine {
	// Инициализация logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Инициализация репозиториев
	expenseRepo := repository.NewExpenseRepository(database.DB, logger)
	budgetRepo := repository.NewBudgetRepository(database.DB, logger)

	// Инициализация сервисов
	expenseService := services.NewExpenseService(expenseRepo, logger)
	budgetService := services.NewBudgetService(budgetRepo, expenseRepo, logger)

	// Инициализация handler'ов
	expenseHandler := handlers.NewExpenseHandler(expenseService, logger)
	budgetHandler := handlers.NewBudgetHandler(budgetService, logger)

	// Настройка роутера
	router := gin.Default()

	handlers.RegisterRoutes(router, database.DB, logger, cfg)

	logger.Info("starting server", slog.String("address", cfg.ServerAddress))

	if err := router.Run(cfg.ServerAddress); err != nil {
		logger.Error("failed to start server", slog.String("error", err.Error()))
		panic(err)
	}
}

// initLogger инициализирует структурированный логгер
func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
	// Регистрация роутов handler'ов
	expenseHandler.RegisterRoutes(router)
	budgetHandler.RegisterRoutes(router)

	return router
}
