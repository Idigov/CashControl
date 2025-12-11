package main

import (
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

	if err := database.Init(cfg); err != nil {
		panic(err)
	}

	if err := database.Migrate(); err != nil {
		panic(err)
	}

	defer database.Close()

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

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Сервер работает",
		})
	})

	// Регистрация роутов handler'ов
	expenseHandler.RegisterRoutes(router)
	budgetHandler.RegisterRoutes(router)

	return router
}
