package handlers

import (
	"cashcontrol/internal/config"
	"cashcontrol/internal/repository"
	"cashcontrol/internal/services"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, logger *slog.Logger, cfg *config.Config) {
	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db, logger)
	categoryRepo := repository.NewCategoryRepository(db, logger)
	expenseRepo := repository.NewExpenseRepository(db, logger)
	budgetRepo := repository.NewBudgetRepository(db, logger)
	recurringExpenseRepo := repository.NewRecurringExpenseRepository(db, logger)
	_ = repository.NewActivityLogRepository(db, logger)
	_ = repository.NewRecurringExpenseRepository(db, logger)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo, logger)
	categoryService := services.NewCategoryService(categoryRepo, logger)
	expenseService := services.NewExpenseService(expenseRepo, logger)
	_ = services.NewBudgetService(budgetRepo, expenseRepo, logger)
	recurringExpenseService := services.NewRecurringExpenseService(recurringExpenseRepo, expenseRepo, logger)

	// Инициализация handlers и регистрация маршрутов
	userHandler := NewUserHandler(userService, logger)
	userHandler.RegisterRoutes(r)

	categoryHandler := NewCategoryHandler(categoryService, logger)
	categoryHandler.RegisterRoutes(r)

	expenseHandler := NewExpenseHandler(expenseService, logger)
	expenseHandler.RegisterRoutes(r)

	recurringExpenseHandler := NewRecurringExpenseHandler(recurringExpenseService, logger)
	recurringExpenseHandler.RegisterRoutes(r)

	// Auth
	authService := services.NewAuthService(userRepo, logger, cfg.JWTSecret)
	authHandler := NewAuthHandler(authService, logger)
	authHandler.RegisterRoutes(r)

	// TODO: Добавить handlers для Budget, ActivityLog, RecurringExpense если необходимо
}
