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

var ErrRecurringExpenseNotFound = errors.New("регулярный расход не найден")

type RecurringExpenseService interface {
	CreateRecurringExpense(userID uint, req models.CreateRecurringExpenseRequest) (*models.RecurringExpense, error)
	GetRecurringExpenseList(userID uint) ([]models.RecurringExpense, error)
	GetRecurringExpenseByID(id uint) (*models.RecurringExpense, error)
	GetActiveRecurringExpenses(userID uint) ([]models.RecurringExpense, error)
	UpdateRecurringExpense(id uint, req models.UpdateRecurringExpenseRequest) (*models.RecurringExpense, error)
	DeleteRecurringExpense(id uint) error
	ActivateRecurringExpense(id uint) (*models.RecurringExpense, error)
	DeactivateRecurringExpense(id uint) (*models.RecurringExpense, error)
	ProcessRecurringExpenses() error
	CalculateNextDate(recurringExpense *models.RecurringExpense) time.Time
}

type recurringExpenseService struct {
	recurringExpenses repository.RecurringExpenseRepository
	expenses          repository.ExpenseRepository
	notifier          NotificationService
	logger            *slog.Logger
}

func NewRecurringExpenseService(
	recurringExpenses repository.RecurringExpenseRepository,
	expenses repository.ExpenseRepository,
	notifier NotificationService,
	logger *slog.Logger,
) RecurringExpenseService {
	return &recurringExpenseService{
		recurringExpenses: recurringExpenses,
		expenses:          expenses,
		notifier:          notifier,
		logger:            logger,
	}
}

func (s *recurringExpenseService) CreateRecurringExpense(userID uint, req models.CreateRecurringExpenseRequest) (*models.RecurringExpense, error) {
	if err := s.validateRecurringExpenseCreate(req); err != nil {
		s.logger.Warn("recurring expense create validation failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.String("type", string(req.Type)),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	// Вычисляем следующую дату создания расхода
	nextDate := s.calculateInitialNextDate(req.Type, req.DayOfWeek, req.DayOfMonth)

	recurringExpense := &models.RecurringExpense{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Description: req.Description,
		Type:        req.Type,
		DayOfWeek:   req.DayOfWeek,
		DayOfMonth:  req.DayOfMonth,
		IsActive:    true,
		NextDate:    nextDate,
	}

	if err := s.recurringExpenses.Create(recurringExpense); err != nil {
		s.logger.Error("recurring expense create failed",
			slog.String("op", "create_recurring_expense"),
			slog.Uint64("user_id", uint64(userID)),
			slog.Any("request", req),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expense created",
		slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
		slog.Uint64("user_id", uint64(userID)),
		slog.String("type", string(req.Type)),
		slog.Time("next_date", nextDate),
	)

	if s.notifier != nil {
		go func() {
			msg := fmt.Sprintf("🔁 Создан регулярный расход: %.2f ₽ (%s). Следующая дата: %s",
				recurringExpense.Amount,
				recurringExpense.Description,
				recurringExpense.NextDate.Format("02.01.2006"),
			)
			if err := s.notifier.SendToUser(userID, msg); err != nil {
				s.logger.Warn("send recurring create notification failed",
					slog.Uint64("user_id", uint64(userID)),
					slog.String("error", err.Error()),
				)
			}
		}()
	}

	return recurringExpense, nil
}

func (s *recurringExpenseService) GetRecurringExpenseList(userID uint) ([]models.RecurringExpense, error) {
	recurringExpenses, err := s.recurringExpenses.GetByUserID(userID)
	if err != nil {
		s.logger.Error("failed to list recurring expenses",
			slog.String("op", "list_recurring_expenses"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expenses listed",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(recurringExpenses)),
	)

	return recurringExpenses, nil
}

func (s *recurringExpenseService) GetRecurringExpenseByID(id uint) (*models.RecurringExpense, error) {
	recurringExpense, err := s.recurringExpenses.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("recurring expense not found",
				slog.Uint64("recurring_expense_id", uint64(id)),
			)
			return nil, ErrRecurringExpenseNotFound
		}
		s.logger.Error("failed to get recurring expense",
			slog.String("op", "get_recurring_expense_by_id"),
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expense retrieved",
		slog.Uint64("recurring_expense_id", uint64(id)),
	)

	return recurringExpense, nil
}

func (s *recurringExpenseService) GetActiveRecurringExpenses(userID uint) ([]models.RecurringExpense, error) {
	allRecurringExpenses, err := s.recurringExpenses.GetByUserID(userID)
	if err != nil {
		s.logger.Error("failed to get recurring expenses",
			slog.String("op", "get_active_recurring_expenses"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	var activeRecurringExpenses []models.RecurringExpense
	for _, re := range allRecurringExpenses {
		if re.IsActive {
			activeRecurringExpenses = append(activeRecurringExpenses, re)
		}
	}

	s.logger.Info("active recurring expenses retrieved",
		slog.Uint64("user_id", uint64(userID)),
		slog.Int("count", len(activeRecurringExpenses)),
	)

	return activeRecurringExpenses, nil
}

func (s *recurringExpenseService) UpdateRecurringExpense(id uint, req models.UpdateRecurringExpenseRequest) (*models.RecurringExpense, error) {
	recurringExpense, err := s.recurringExpenses.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("recurring expense not found for update",
				slog.Uint64("recurring_expense_id", uint64(id)),
			)
			return nil, ErrRecurringExpenseNotFound
		}
		s.logger.Error("failed to fetch recurring expense before update",
			slog.String("op", "update_recurring_expense"),
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := s.applyRecurringExpenseUpdate(recurringExpense, req); err != nil {
		s.logger.Warn("recurring expense update validation failed",
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.Any("request", req),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	// Пересчитываем следующую дату, если изменился тип или параметры
	if req.Type != nil || req.DayOfWeek != nil || req.DayOfMonth != nil {
		recurringExpense.NextDate = s.CalculateNextDate(recurringExpense)
	}

	if err := s.recurringExpenses.Update(recurringExpense); err != nil {
		s.logger.Error("recurring expense update failed",
			slog.String("op", "update_recurring_expense"),
			slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expense updated",
		slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
	)

	return recurringExpense, nil
}

func (s *recurringExpenseService) DeleteRecurringExpense(id uint) error {
	_, err := s.recurringExpenses.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("recurring expense not found for delete",
				slog.Uint64("recurring_expense_id", uint64(id)),
			)
			return ErrRecurringExpenseNotFound
		}
		s.logger.Error("failed to fetch recurring expense before delete",
			slog.String("op", "delete_recurring_expense"),
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.recurringExpenses.Delete(id); err != nil {
		s.logger.Error("recurring expense delete failed",
			slog.String("op", "delete_recurring_expense"),
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("recurring expense deleted",
		slog.Uint64("recurring_expense_id", uint64(id)),
	)

	return nil
}

func (s *recurringExpenseService) ActivateRecurringExpense(id uint) (*models.RecurringExpense, error) {
	recurringExpense, err := s.recurringExpenses.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecurringExpenseNotFound
		}
		return nil, err
	}

	recurringExpense.IsActive = true
	if err := s.recurringExpenses.Update(recurringExpense); err != nil {
		s.logger.Error("failed to activate recurring expense",
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expense activated",
		slog.Uint64("recurring_expense_id", uint64(id)),
	)

	if s.notifier != nil {
		go func() {
			msg := fmt.Sprintf("✅ Регулярный расход включен: %.2f ₽ (%s). Следующая дата: %s",
				recurringExpense.Amount,
				recurringExpense.Description,
				recurringExpense.NextDate.Format("02.01.2006"),
			)
			if err := s.notifier.SendToUser(recurringExpense.UserID, msg); err != nil {
				s.logger.Warn("send recurring activate notification failed",
					slog.Uint64("user_id", uint64(recurringExpense.UserID)),
					slog.String("error", err.Error()),
				)
			}
		}()
	}

	return recurringExpense, nil
}

func (s *recurringExpenseService) DeactivateRecurringExpense(id uint) (*models.RecurringExpense, error) {
	recurringExpense, err := s.recurringExpenses.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecurringExpenseNotFound
		}
		return nil, err
	}

	recurringExpense.IsActive = false
	if err := s.recurringExpenses.Update(recurringExpense); err != nil {
		s.logger.Error("failed to deactivate recurring expense",
			slog.Uint64("recurring_expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("recurring expense deactivated",
		slog.Uint64("recurring_expense_id", uint64(id)),
	)

	return recurringExpense, nil
}

func (s *recurringExpenseService) ProcessRecurringExpenses() error {
	now := time.Now()
	dueRecurringExpenses, err := s.recurringExpenses.GetActiveByNextDate(now)
	if err != nil {
		s.logger.Error("failed to get due recurring expenses",
			slog.String("op", "process_recurring_expenses"),
			slog.String("error", err.Error()),
		)
		return err
	}

	for _, recurringExpense := range dueRecurringExpenses {
		err = repository.RunInTransaction(func(tx repository.TxProvider) error {
			// Создаем расход
			expense := &models.Expense{
				UserID:      recurringExpense.UserID,
				CategoryID:  recurringExpense.CategoryID,
				Amount:      recurringExpense.Amount,
				Description: recurringExpense.Description,
				Date:        recurringExpense.NextDate,
			}

			if err := s.expenses.WithTx(tx).Create(expense); err != nil {
				return fmt.Errorf("create expense from recurring %d: %w", recurringExpense.ID, err)
			}

			// Обновляем следующую дату
			recurringExpense.NextDate = s.CalculateNextDate(&recurringExpense)
			if err := s.recurringExpenses.WithTx(tx).Update(&recurringExpense); err != nil {
				return fmt.Errorf("update next_date for recurring %d: %w", recurringExpense.ID, err)
			}

			// Уведомление после commit
			if s.notifier != nil {
				msg := fmt.Sprintf("🔁 Сегодня списание: %.2f ₽ (%s)%s",
					recurringExpense.Amount,
					recurringExpense.Description,
					func() string {
						if recurringExpense.Category.Name != "" {
							return fmt.Sprintf(" — категория %s", recurringExpense.Category.Name)
						}
						return ""
					}(),
				)
				defer func(re models.RecurringExpense) {
					go func() {
						if err := s.notifier.SendToUser(re.UserID, msg); err != nil {
							s.logger.Warn("send recurring due notification failed",
								slog.Uint64("user_id", uint64(re.UserID)),
								slog.String("error", err.Error()),
							)
						}
					}()
				}(recurringExpense)
			}

			s.logger.Info("processed recurring expense",
				slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
				slog.Float64("amount", recurringExpense.Amount),
				slog.Time("next_date", recurringExpense.NextDate),
			)
			return nil
		})

		if err != nil {
			s.logger.Error("process recurring expense failed",
				slog.Uint64("recurring_expense_id", uint64(recurringExpense.ID)),
				slog.String("error", err.Error()),
			)
			continue
		}
	}

	return nil
}

func (s *recurringExpenseService) CalculateNextDate(recurringExpense *models.RecurringExpense) time.Time {
	now := time.Now()
	baseDate := recurringExpense.NextDate
	if baseDate.Before(now) {
		baseDate = now
	}

	switch recurringExpense.Type {
	case models.RecurringTypeDaily:
		return baseDate.AddDate(0, 0, 1)

	case models.RecurringTypeWeekly:
		if recurringExpense.DayOfWeek != nil {
			// Находим следующий указанный день недели
			daysUntilTarget := (int(*recurringExpense.DayOfWeek) - int(baseDate.Weekday()) + 7) % 7
			if daysUntilTarget == 0 {
				daysUntilTarget = 7 // Если сегодня нужный день, берем следующий
			}
			return baseDate.AddDate(0, 0, daysUntilTarget)
		}
		return baseDate.AddDate(0, 0, 7)

	case models.RecurringTypeMonthly:
		if recurringExpense.DayOfMonth != nil {
			// Находим следующий месяц с указанным днем
			nextMonth := baseDate.AddDate(0, 1, 0)
			day := *recurringExpense.DayOfMonth
			// Проверяем, что день существует в месяце
			lastDayOfMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, nextMonth.Location()).Day()
			if day > lastDayOfMonth {
				day = lastDayOfMonth
			}
			return time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, nextMonth.Location())
		}
		return baseDate.AddDate(0, 1, 0)

	case models.RecurringTypeYearly:
		return baseDate.AddDate(1, 0, 0)

	default:
		return baseDate.AddDate(0, 0, 1)
	}
}

func (s *recurringExpenseService) validateRecurringExpenseCreate(req models.CreateRecurringExpenseRequest) error {
	if req.Amount <= 0 {
		return errors.New("сумма должна быть больше нуля")
	}

	// Валидация типа повторения
	switch req.Type {
	case models.RecurringTypeDaily:
		// Для ежедневных не нужны дополнительные параметры
		return nil

	case models.RecurringTypeWeekly:
		if req.DayOfWeek == nil {
			return errors.New("для еженедельных расходов необходимо указать день недели")
		}
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			return errors.New("день недели должен быть от 0 (воскресенье) до 6 (суббота)")
		}
		return nil

	case models.RecurringTypeMonthly:
		if req.DayOfMonth == nil {
			return errors.New("для ежемесячных расходов необходимо указать день месяца")
		}
		if *req.DayOfMonth < 1 || *req.DayOfMonth > 31 {
			return errors.New("день месяца должен быть от 1 до 31")
		}
		return nil

	case models.RecurringTypeYearly:
		// Для ежегодных не нужны дополнительные параметры
		return nil

	default:
		return errors.New("неподдерживаемый тип повторения")
	}
}

func (s *recurringExpenseService) applyRecurringExpenseUpdate(
	recurringExpense *models.RecurringExpense,
	req models.UpdateRecurringExpenseRequest,
) error {
	if req.CategoryID != nil {
		recurringExpense.CategoryID = *req.CategoryID
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return errors.New("сумма должна быть больше нуля")
		}
		recurringExpense.Amount = *req.Amount
	}

	if req.Description != nil {
		recurringExpense.Description = *req.Description
	}

	if req.Type != nil {
		recurringExpense.Type = *req.Type
	}

	if req.DayOfWeek != nil {
		if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			return errors.New("день недели должен быть от 0 (воскресенье) до 6 (суббота)")
		}
		recurringExpense.DayOfWeek = req.DayOfWeek
	}

	if req.DayOfMonth != nil {
		if *req.DayOfMonth < 1 || *req.DayOfMonth > 31 {
			return errors.New("день месяца должен быть от 1 до 31")
		}
		recurringExpense.DayOfMonth = req.DayOfMonth
	}

	if req.IsActive != nil {
		recurringExpense.IsActive = *req.IsActive
	}

	return nil
}

func (s *recurringExpenseService) calculateInitialNextDate(
	expenseType models.RecurringExpenseType,
	dayOfWeek *int,
	dayOfMonth *int,
) time.Time {
	now := time.Now()

	switch expenseType {
	case models.RecurringTypeDaily:
		return now.AddDate(0, 0, 1)

	case models.RecurringTypeWeekly:
		if dayOfWeek != nil {
			daysUntilTarget := (int(*dayOfWeek) - int(now.Weekday()) + 7) % 7
			if daysUntilTarget == 0 {
				daysUntilTarget = 7
			}
			return now.AddDate(0, 0, daysUntilTarget)
		}
		return now.AddDate(0, 0, 7)

	case models.RecurringTypeMonthly:
		if dayOfMonth != nil {
			day := *dayOfMonth
			// Если день месяца уже прошел в текущем месяце, берем следующий месяц
			if day < now.Day() {
				nextMonth := now.AddDate(0, 1, 0)
				lastDayOfMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, nextMonth.Location()).Day()
				if day > lastDayOfMonth {
					day = lastDayOfMonth
				}
				return time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, nextMonth.Location())
			}
			// Иначе берем текущий месяц
			lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
			if day > lastDayOfMonth {
				day = lastDayOfMonth
			}
			return time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		}
		return now.AddDate(0, 1, 0)

	case models.RecurringTypeYearly:
		return now.AddDate(1, 0, 0)

	default:
		return now.AddDate(0, 0, 1)
	}
}
