package handlers

import (
	"cashcontrol/internal/config"
	"cashcontrol/internal/middleware"
	"cashcontrol/internal/repository"
	"cashcontrol/internal/services"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, logger *slog.Logger, cfg *config.Config) {
	// ---------- repositories ----------
	userRepo := repository.NewUserRepository(db, logger)
	categoryRepo := repository.NewCategoryRepository(db, logger)
	expenseRepo := repository.NewExpenseRepository(db, logger)
	budgetRepo := repository.NewBudgetRepository(db, logger)
	recurringExpenseRepo := repository.NewRecurringExpenseRepository(db, logger)
	_ = repository.NewActivityLogRepository(db, logger)
	_ = repository.NewRecurringExpenseRepository(db, logger)

	// ---------- services ----------
	userService := services.NewUserService(userRepo, logger)
	categoryService := services.NewCategoryService(categoryRepo, logger)
	expenseService := services.NewExpenseService(expenseRepo, categoryRepo, logger)
	budgetService := services.NewBudgetService(budgetRepo, expenseRepo, logger)
	recurringExpenseService := services.NewRecurringExpenseService(recurringExpenseRepo, expenseRepo, logger)

	// ---------- API root ----------
	api := r.Group("/api")

	// ---------- AUTH (PUBLIC) ----------
	authService := services.NewAuthService(
		userRepo,
		categoryRepo,
		logger,
		cfg.JWTSecret,
		cfg.TelegramBotToken,
	)
	authHandler := NewAuthHandler(authService, logger)
	authHandler.RegisterRoutes(api, cfg.JWTSecret)

	// ---------- PROTECTED ----------
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// ---------- handlers ----------
	userHandler := NewUserHandler(userService, logger)
	userHandler.RegisterRoutes(protected)

	categoryHandler := NewCategoryHandler(categoryService, logger)
	categoryHandler.RegisterRoutes(protected)

	expenseHandler := NewExpenseHandler(expenseService, logger)
	expenseHandler.RegisterRoutes(protected)

	budgetHandler := NewBudgetHandler(budgetService, logger)
	budgetHandler.RegisterRoutes(protected)

	recurringExpenseHandler := NewRecurringExpenseHandler(recurringExpenseService, logger)
	recurringExpenseHandler.RegisterRoutes(protected)

	analyticsRepo := repository.NewAnalyticsRepository(db)
	analyticsService := services.NewAnalyticsService(analyticsRepo)
	analyticsHandler := NewAnalyticsHandler(analyticsService)
	analyticsHandler.RegisterRoutes(protected)

	statsRepo := repository.NewStatisticsRepository(db)
	statsService := services.NewStatisticsService(statsRepo)
	statsHandler := NewStatisticsHandler(statsService)

	statsHandler.RegisterRoutes(api) 

}
