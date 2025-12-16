package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

var ErrBudgetNotFound = errors.New("бюджет не найден")

const (
	NearLimitThreshold = 0.8 // 80% использования бюджета
)

type BudgetService interface {
	CreateBudget(userID uint, req models.CreateBudgetRequest) (*models.Budget, error)
	GetBudgetList(userID uint) ([]models.Budget, error)
	GetBudgetByID(id uint) (*models.Budget, error)
	GetBudgetByUserIDAndMonth(userID uint, month, year int) (*models.Budget, error)
	GetCurrentBudgetStatus(userID uint) (*models.BudgetStatus, error)
	GetBudgetStatus(userID uint, month, year int) (*models.BudgetStatus, error)
	UpdateBudget(id uint, req models.UpdateBudgetRequest) (*models.Budget, error)
	DeleteBudget(id uint) error
}

type budgetService struct {
	budgets  repository.BudgetRepository
	expenses repository.ExpenseRepository
	notifier NotificationService
	logger   *slog.Logger
}

func NewBudgetService(
	budgets repository.BudgetRepository,
	expenses repository.ExpenseRepository,
	notifier NotificationService,
	logger *slog.Logger,
) BudgetService {
	return &budgetService{
		budgets:  budgets,
		expenses: expenses,
		notifier: notifier,
		logger:   logger,
	}
}

func (s *budgetService) CreateBudget(userID uint, req models.CreateBudgetRequest) (*models.Budget, error) {
	if err := s.validateBudgetCreate(req); err != nil {
		s.logger.Warn("budget create validation failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.Float64("amount", req.Amount),
			slog.Int("month", req.Month),
			slog.Int("year", req.Year),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	// Проверяем, не существует ли уже бюджет на этот месяц
	existing, err := s.budgets.GetByUserIDAndMonth(userID, req.Month, req.Year)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to check existing budget",
				slog.Uint64("user_id", uint64(userID)),
				slog.Int("month", req.Month),
				slog.Int("year", req.Year),
				slog.String("error", err.Error()),
			)
			return nil, err
		}
		// ErrRecordNotFound - это нормально, бюджета нет
	} else if existing != nil {
		s.logger.Warn("budget already exists for this month",
			slog.Uint64("user_id", uint64(userID)),
			slog.Int("month", req.Month),
			slog.Int("year", req.Year),
		)
		return nil, errors.New("бюджет на этот месяц уже существует")
	}

	budget := &models.Budget{
		UserID: userID,
		Amount: req.Amount,
		Month:  req.Month,
		Year:   req.Year,
	}

	if err := s.budgets.Create(budget); err != nil {
		s.logger.Error("budget create failed",
			slog.String("op", "create_budget"),
			slog.Uint64("user_id", uint64(userID)),
			slog.Any("request", req),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("budget created",
		slog.Uint64("budget_id", uint64(budget.ID)),
		slog.Uint64("user_id", uint64(userID)),
		slog.Float64("amount", budget.Amount),
		slog.Int("month", budget.Month),
		slog.Int("year", budget.Year),
	)

	return budget, nil
}

func (s *budgetService) GetBudgetList(userID uint) ([]models.Budget, error) {
	budgets, err := s.budgets.GetByUserID(userID)
	if err != nil {
		s.logger.Error("failed to list budgets",
			slog.String("op", "list_budgets"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("budgets listed",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(budgets)),
	)

	return budgets, nil
}

func (s *budgetService) GetCurrentBudgetStatus(userID uint) (*models.BudgetStatus, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	return s.GetBudgetStatus(userID, month, year)
}

func (s *budgetService) GetBudgetByID(id uint) (*models.Budget, error) {
	budget, err := s.budgets.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("budget not found",
				slog.Uint64("budget_id", uint64(id)),
			)
			return nil, ErrBudgetNotFound
		}
		s.logger.Error("failed to get budget",
			slog.String("op", "get_budget_by_id"),
			slog.Uint64("budget_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("budget retrieved",
		slog.Uint64("budget_id", uint64(budget.ID)),
		slog.Uint64("user_id", uint64(budget.UserID)),
		slog.Float64("amount", budget.Amount),
	)

	return budget, nil
}

func (s *budgetService) GetBudgetByUserIDAndMonth(userID uint, month, year int) (*models.Budget, error) {
	budget, err := s.budgets.GetByUserIDAndMonth(userID, month, year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("budget not found",
				slog.Uint64("user_id", uint64(userID)),
				slog.Int("month", month),
				slog.Int("year", year),
			)
			return nil, ErrBudgetNotFound
		}
		s.logger.Error("failed to get budget",
			slog.String("op", "get_budget_by_user_id_and_month"),
			slog.Uint64("user_id", uint64(userID)),
			slog.Int("month", month),
			slog.Int("year", year),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return budget, nil
}

func (s *budgetService) GetBudgetStatus(userID uint, month, year int) (*models.BudgetStatus, error) {
	budget, err := s.budgets.GetByUserIDAndMonth(userID, month, year)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}

	// Расчет потраченной суммы за период
	spent, err := s.calculateSpentAmount(userID, month, year)
	if err != nil {
		s.logger.Error("failed to calculate spent amount",
			slog.String("op", "get_budget_status"),
			slog.Uint64("user_id", uint64(userID)),
			slog.Int("month", month),
			slog.Int("year", year),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Расчет оставшегося лимита
	remaining := budget.Amount - spent
	if remaining < 0 {
		remaining = 0
	}

	// Определение процента использования бюджета
	var percentage float64
	if budget.Amount == 0 {
		percentage = 0
	} else {
		percentage = (spent / budget.Amount) * 100
	}

	// Проверка превышения и приближения к лимиту
	isExceeded := spent > budget.Amount
	isNearLimit := percentage >= (NearLimitThreshold*100) && !isExceeded

	status := &models.BudgetStatus{
		Budget:      budget,
		Spent:       spent,
		Remaining:   remaining,
		Percentage:  percentage,
		IsExceeded:  isExceeded,
		IsNearLimit: isNearLimit,
	}

	// Логирование уведомлений
	if isExceeded {
		s.logger.Warn("budget exceeded",
			slog.Uint64("budget_id", uint64(budget.ID)),
			slog.Uint64("user_id", uint64(userID)),
			slog.Float64("budget_amount", budget.Amount),
			slog.Float64("spent", spent),
			slog.Float64("percentage", percentage),
		)
	} else if isNearLimit {
		s.logger.Info("budget near limit",
			slog.Uint64("budget_id", uint64(budget.ID)),
			slog.Uint64("user_id", uint64(userID)),
			slog.Float64("budget_amount", budget.Amount),
			slog.Float64("spent", spent),
			slog.Float64("percentage", percentage),
		)
	}

	// Уведомления
	if s.notifier != nil && (isExceeded || isNearLimit) {
		go func() {
			var msg string
			if isExceeded {
				msg = fmt.Sprintf("🚨 Бюджет превышен: потрачено %.0f%% (%.0f / %.0f)", percentage, spent, budget.Amount)
			} else {
				msg = fmt.Sprintf("⚠️ Бюджет на исходе: %.0f%% (%.0f / %.0f)", percentage, spent, budget.Amount)
			}
			if err := s.notifier.SendToUser(userID, msg); err != nil {
				s.logger.Warn("send budget notification failed", slog.Uint64("user_id", uint64(userID)), slog.String("error", err.Error()))
			}
		}()
	}

	return status, nil
}

func (s *budgetService) UpdateBudget(id uint, req models.UpdateBudgetRequest) (*models.Budget, error) {
	budget, err := s.budgets.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("budget not found for update",
				slog.Uint64("budget_id", uint64(id)),
			)
			return nil, ErrBudgetNotFound
		}
		s.logger.Error("failed to fetch budget before update",
			slog.String("op", "update_budget"),
			slog.Uint64("budget_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := s.applyBudgetUpdate(budget, req); err != nil {
		s.logger.Warn("budget update validation failed",
			slog.Uint64("budget_id", uint64(id)),
			slog.Any("request", req),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	if err := s.budgets.Update(budget); err != nil {
		s.logger.Error("budget update failed",
			slog.String("op", "update_budget"),
			slog.Uint64("budget_id", uint64(budget.ID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("budget updated",
		slog.Uint64("budget_id", uint64(budget.ID)),
		slog.Float64("amount", budget.Amount),
		slog.Int("month", budget.Month),
		slog.Int("year", budget.Year),
	)

	return budget, nil
}

func (s *budgetService) DeleteBudget(id uint) error {
	_, err := s.budgets.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("budget not found for delete",
				slog.Uint64("budget_id", uint64(id)),
			)
			return ErrBudgetNotFound
		}
		s.logger.Error("failed to fetch budget before delete",
			slog.String("op", "delete_budget"),
			slog.Uint64("budget_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.budgets.Delete(id); err != nil {
		s.logger.Error("budget delete failed",
			slog.String("op", "delete_budget"),
			slog.Uint64("budget_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("budget deleted",
		slog.Uint64("budget_id", uint64(id)),
	)

	return nil
}

func (s *budgetService) validateBudgetCreate(req models.CreateBudgetRequest) error {
	if req.Amount <= 0 {
		return errors.New("сумма бюджета должна быть больше нуля")
	}

	if req.Month < 1 || req.Month > 12 {
		return errors.New("месяц должен быть от 1 до 12")
	}

	if req.Year < 2000 || req.Year > 2100 {
		return errors.New("год должен быть в диапазоне 2000-2100")
	}

	return nil
}

func (s *budgetService) applyBudgetUpdate(budget *models.Budget, req models.UpdateBudgetRequest) error {
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return errors.New("сумма бюджета должна быть больше нуля")
		}
		budget.Amount = *req.Amount
	}

	if req.Month != nil {
		if *req.Month < 1 || *req.Month > 12 {
			return errors.New("месяц должен быть от 1 до 12")
		}
		budget.Month = *req.Month
	}

	if req.Year != nil {
		if *req.Year < 2000 || *req.Year > 2100 {
			return errors.New("год должен быть в диапазоне 2000-2100")
		}
		budget.Year = *req.Year
	}

	return nil
}

func (s *budgetService) calculateSpentAmount(userID uint, month, year int) (float64, error) {
	// Определяем начало и конец месяца
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Получаем расходы пользователя за указанный период
	filter := models.ExpenseFilter{
		UserID:    userID,
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	expenses, err := s.expenses.List(filter)
	if err != nil {
		return 0, err
	}

	// Суммируем расходы
	var total float64
	for _, expense := range expenses {
		total += expense.Amount
	}

	return total, nil
}
