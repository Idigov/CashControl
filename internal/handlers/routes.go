package handlers

import (
	"cashcontrol/internal/config"
	"cashcontrol/internal/middleware"
	"cashcontrol/internal/repository"
	"cashcontrol/internal/services"
	"log/slog"
	"time"

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
	notificationService, err := services.NewNotificationService(cfg.TelegramBotToken, userRepo, logger)
	if err != nil {
		logger.Warn("notification service init failed", slog.String("error", err.Error()))
	}

	budgetService := services.NewBudgetService(budgetRepo, expenseRepo, notificationService, logger)
	recurringExpenseService := services.NewRecurringExpenseService(recurringExpenseRepo, expenseRepo, notificationService, logger)

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
	analyticsService := services.NewAnalyticsService(analyticsRepo, logger)
	analyticsHandler := NewAnalyticsHandler(analyticsService, logger)
	analyticsHandler.RegisterRoutes(protected)

	statsRepo := repository.NewStatisticsRepository(db)
	statsService := services.NewStatisticsService(statsRepo, logger)
	statsHandler := NewStatisticsHandler(statsService, logger)
	statsHandler.RegisterRoutes(protected) 

	// Напоминание записывать расходы (каждый день)
	if notificationService != nil {
		go startDailyExpenseReminder(notificationService, userRepo, logger)
		go startRecurringProcessor(recurringExpenseService, logger)
	}

}

func startDailyExpenseReminder(notification services.NotificationService, users repository.UserRepository, logger *slog.Logger) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		time.Sleep(next.Sub(now))

		list, err := users.List()
		if err != nil {
			logger.Warn("daily reminder: failed to list users", slog.String("error", err.Error()))
			continue
		}

		for _, u := range list {
			var chatID int64
			if u.TelegramChatID != nil {
				chatID = *u.TelegramChatID
			} else if u.TelegramID != nil {
				chatID = *u.TelegramID
			}
			if chatID == 0 {
				continue
			}
			go func(id int64) {
				msg := "🧾 Не забудьте записать сегодняшние расходы в CashControl"
				if err := notification.SendToChat(id, msg); err != nil {
					logger.Warn("daily reminder send failed", slog.Int64("chat_id", id), slog.String("error", err.Error()))
				}
			}(chatID)
		}
	}
}

func startRecurringProcessor(recurring services.RecurringExpenseService, logger *slog.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		if err := recurring.ProcessRecurringExpenses(); err != nil {
			logger.Warn("process recurring expenses failed", slog.String("error", err.Error()))
		}
		<-ticker.C
	}
}
