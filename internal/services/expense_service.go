package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var ErrExpenseNotFound = errors.New("расход не найден")

type ExpenseService interface {
	CreateExpense(userID uint, req models.CreateExpenseRequest) (*models.Expense, error)
	GetExpenseList(filter models.ExpenseFilter) ([]models.Expense, error)
	GetExpenseByID(id uint) (*models.Expense, error)
	UpdateExpense(id uint, req models.UpdateExpenseRequest) (*models.Expense, error)
	DeleteExpense(id uint) error
}

type expenseService struct {
	expenses   repository.ExpenseRepository
	categories repository.CategoryRepository
	logger     *slog.Logger
}

func NewExpenseService(expenses repository.ExpenseRepository, categories repository.CategoryRepository, logger *slog.Logger) ExpenseService {
	return &expenseService{
		expenses:   expenses,
		categories: categories,
		logger:     logger,
	}
}

func (s *expenseService) CreateExpense(userID uint, req models.CreateExpenseRequest) (*models.Expense, error) {

	if err := s.validateExpenseCreate(req); err != nil {
		s.logger.Warn("expense create validation failed",
			slog.Uint64("user_id", uint64(userID)),
			slog.Uint64("category_id", uint64(req.CategoryID)),
			slog.Float64("amount", req.Amount),
			slog.Time("date", req.Date),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	expense := &models.Expense{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		Date:        req.Date,
		Amount:      req.Amount,
	}
	if err := s.expenses.Create(expense); err != nil {
		s.logger.Error("expense create failed",
			slog.String("op", "create_expense"),
			slog.Any("request", req),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	s.logger.Info("expense created",
		slog.Uint64("expense_id", uint64(expense.ID)),
		slog.Uint64("user_id", uint64(userID)),
		slog.Uint64("category_id", uint64(expense.CategoryID)),
		slog.Float64("amount", expense.Amount),
		slog.Time("date", expense.Date),
	)
	return expense, nil
}

func (s *expenseService) GetExpenseList(filter models.ExpenseFilter) ([]models.Expense, error) {
	expenses, err := s.expenses.List(filter)

	if err != nil {
		s.logger.Error("failed to list expenses",
			slog.String("op", "list_expenses"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("expenses listed",
		slog.Int("count", len(expenses)),
	)

	return expenses, nil
}

func (s *expenseService) GetExpenseByID(id uint) (*models.Expense, error) {
	expense, err := s.expenses.GetByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("expense not found",
				slog.Uint64("expense_id", uint64(id)),
			)
			return nil, ErrExpenseNotFound
		}
		s.logger.Error("failed to get expense",
			slog.String("op", "get_expense_by_id"),
			slog.Uint64("expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("expense retrieved",
		slog.Uint64("expense_id", uint64(expense.ID)),
		slog.Uint64("category_id", uint64(expense.CategoryID)),
		slog.Float64("amount", expense.Amount),
	)

	return expense, nil
}

func (s *expenseService) UpdateExpense(id uint, req models.UpdateExpenseRequest) (*models.Expense, error) {
	expense, err := s.expenses.GetByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("expense not found for update",
				slog.Uint64("expense_id", uint64(id)),
			)
			return nil, ErrExpenseNotFound
		}
		s.logger.Error("failed to fetch expense before update",
			slog.String("op", "update_expense"),
			slog.Uint64("expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if err := s.applyExpenseUpdate(expense, req); err != nil {
		s.logger.Warn("expense update validation failed",
			slog.Uint64("expense_id", uint64(id)),
			slog.Any("request", req),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	if err := s.expenses.Update(expense); err != nil {
		s.logger.Error("expense update failed",
			slog.String("op", "update_expense"),
			slog.Uint64("expense_id", uint64(expense.ID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	s.logger.Info("expense updated",
		slog.Uint64("expense_id", uint64(expense.ID)),
		slog.Uint64("category_id", uint64(expense.CategoryID)),
		slog.Float64("amount", expense.Amount),
		slog.Time("date", expense.Date),
	)
	return expense, nil
}

func (s *expenseService) DeleteExpense(id uint) error {
	_, err := s.expenses.GetByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("expense not found for delete",
				slog.Uint64("expense_id", uint64(id)),
			)
			return ErrExpenseNotFound
		}
		s.logger.Error("failed to fetch expense before delete",
			slog.String("op", "delete_expense"),
			slog.Uint64("expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.expenses.Delete(id); err != nil {
		s.logger.Error("expense delete failed",
			slog.String("op", "delete_expense"),
			slog.Uint64("expense_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("expense deleted",
		slog.Uint64("expense_id", uint64(id)),
	)

	return nil
}

func (s *expenseService) validateExpenseCreate(req models.CreateExpenseRequest) error {
	if req.Amount <= 0 {
		return errors.New("сумма должна быть больше нуля")
	}

	// Проверка существования категории
	_, err := s.categories.GetByID(req.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("категория не найдена")
		}
		s.logger.Error("failed to check category existence",
			slog.Uint64("category_id", uint64(req.CategoryID)),
			slog.String("error", err.Error()),
		)
		return errors.New("ошибка при проверке категории")
	}

	return nil
}

func (s *expenseService) applyExpenseUpdate(expense *models.Expense, req models.UpdateExpenseRequest) error {
	if req.CategoryID != nil {
		expense.CategoryID = *req.CategoryID
		// TODO: добавить валидацию существования категории при обновлении
	}

	if req.Description != nil {
		expense.Description = *req.Description
	}

	if req.Amount != nil {
		if *req.Amount <= 0 {
			return errors.New("сумма должна быть больше нуля")
		}
		expense.Amount = *req.Amount
	}

	if req.Date != nil {
		expense.Date = *req.Date
	}

	return nil
}
