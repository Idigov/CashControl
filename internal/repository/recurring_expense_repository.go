package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

var errRecurringExpenseNil error = errors.New("recurring expense is nil")

type RecurringExpenseRepository interface {
	List() ([]models.RecurringExpense, error)
	GetByID(id uint) (*models.RecurringExpense, error)
	GetByUserID(userID uint) ([]models.RecurringExpense, error)
	GetActiveByNextDate(nextDate time.Time) ([]models.RecurringExpense, error)
	Create(recurringExpense *models.RecurringExpense) error
	Update(recurringExpense *models.RecurringExpense) error
	Delete(id uint) error
	WithTx(tx TxProvider) RecurringExpenseRepository
}

type gormRecurringExpenseRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewRecurringExpenseRepository(db *gorm.DB, logger *slog.Logger) RecurringExpenseRepository {
	return &gormRecurringExpenseRepository{db: db, logger: logger}
}

func (r *gormRecurringExpenseRepository) WithTx(tx TxProvider) RecurringExpenseRepository {
	return &gormRecurringExpenseRepository{db: tx.DB(), logger: r.logger}
}

func (r *gormRecurringExpenseRepository) List() ([]models.RecurringExpense, error) {
	r.logger.Debug("repo.recurring_expense.list",
		slog.String("op", "repo.recurring_expense.list"),
	)
	var recurringExpenses []models.RecurringExpense
	if err := r.db.Find(&recurringExpenses).Error; err != nil {
		r.logger.Error("repo.recurring_expense.list failed",
			slog.String("op", "repo.recurring_expense.list"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return recurringExpenses, nil
}

func (r *gormRecurringExpenseRepository) GetByID(id uint) (*models.RecurringExpense, error) {
	r.logger.Debug("repo.recurring_expense.get_by_id",
		slog.String("op", "repo.recurring_expense.get_by_id"),
		slog.Uint64("id", uint64(id)),
	)
	var recurringExpense models.RecurringExpense
	if err := r.db.Preload("Category").First(&recurringExpense, id).Error; err != nil {
		r.logger.Error("repo.recurring_expense.get_by_id failed",
			slog.String("op", "repo.recurring_expense.get_by_id"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &recurringExpense, nil
}

func (r *gormRecurringExpenseRepository) GetByUserID(userID uint) ([]models.RecurringExpense, error) {
	r.logger.Debug("repo.recurring_expense.get_by_user_id",
		slog.String("op", "repo.recurring_expense.get_by_user_id"),
		slog.Uint64("user_id", uint64(userID)),
	)
	var recurringExpenses []models.RecurringExpense
	if err := r.db.Preload("Category").Where("user_id = ?", userID).Order("next_date ASC").Find(&recurringExpenses).Error; err != nil {
		r.logger.Error("repo.recurring_expense.get_by_user_id failed",
			slog.String("op", "repo.recurring_expense.get_by_user_id"),
			slog.Uint64("user_id", uint64(userID)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return recurringExpenses, nil
}

func (r *gormRecurringExpenseRepository) GetActiveByNextDate(nextDate time.Time) ([]models.RecurringExpense, error) {
	r.logger.Debug("repo.recurring_expense.get_active_by_next_date",
		slog.String("op", "repo.recurring_expense.get_active_by_next_date"),
		slog.Time("next_date", nextDate),
	)
	var recurringExpenses []models.RecurringExpense
	if err := r.db.Preload("Category").Where("is_active = ? AND next_date <= ?", true, nextDate).Find(&recurringExpenses).Error; err != nil {
		r.logger.Error("repo.recurring_expense.get_active_by_next_date failed",
			slog.String("op", "repo.recurring_expense.get_active_by_next_date"),
			slog.Time("next_date", nextDate),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return recurringExpenses, nil
}

func (r *gormRecurringExpenseRepository) Create(recurringExpense *models.RecurringExpense) error {
	if recurringExpense == nil {
		return errRecurringExpenseNil
	}
	r.logger.Debug("repo.recurring_expense.create",
		slog.String("op", "repo.recurring_expense.create"),
		slog.Uint64("user_id", uint64(recurringExpense.UserID)),
		slog.Uint64("category_id", uint64(recurringExpense.CategoryID)),
		slog.Float64("amount", recurringExpense.Amount),
		slog.String("type", string(recurringExpense.Type)),
		slog.Time("next_date", recurringExpense.NextDate),
	)
	if err := r.db.Create(recurringExpense).Error; err != nil {
		r.logger.Error("repo.recurring_expense.create failed",
			slog.String("op", "repo.recurring_expense.create"),
			slog.Uint64("user_id", uint64(recurringExpense.UserID)),
			slog.Uint64("category_id", uint64(recurringExpense.CategoryID)),
			slog.Float64("amount", recurringExpense.Amount),
			slog.String("type", string(recurringExpense.Type)),
			slog.Time("next_date", recurringExpense.NextDate),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormRecurringExpenseRepository) Update(recurringExpense *models.RecurringExpense) error {
	if recurringExpense == nil {
		return errRecurringExpenseNil
	}
	r.logger.Debug("repo.recurring_expense.update",
		slog.String("op", "repo.recurring_expense.update"),
		slog.Uint64("id", uint64(recurringExpense.ID)),
		slog.Uint64("user_id", uint64(recurringExpense.UserID)),
		slog.Uint64("category_id", uint64(recurringExpense.CategoryID)),
		slog.Float64("amount", recurringExpense.Amount),
		slog.String("type", string(recurringExpense.Type)),
		slog.Time("next_date", recurringExpense.NextDate),
	)
	if err := r.db.Save(recurringExpense).Error; err != nil {
		r.logger.Error("repo.recurring_expense.update failed",
			slog.String("op", "repo.recurring_expense.update"),
			slog.Uint64("id", uint64(recurringExpense.ID)),
			slog.Uint64("user_id", uint64(recurringExpense.UserID)),
			slog.Uint64("category_id", uint64(recurringExpense.CategoryID)),
			slog.Float64("amount", recurringExpense.Amount),
			slog.String("type", string(recurringExpense.Type)),
			slog.Time("next_date", recurringExpense.NextDate),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormRecurringExpenseRepository) Delete(id uint) error {
	r.logger.Debug("repo.recurring_expense.delete",
		slog.String("op", "repo.recurring_expense.delete"),
		slog.Uint64("id", uint64(id)),
	)
	if err := r.db.Delete(&models.RecurringExpense{}, id).Error; err != nil {
		r.logger.Error("repo.recurring_expense.delete failed",
			slog.String("op", "repo.recurring_expense.delete"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
