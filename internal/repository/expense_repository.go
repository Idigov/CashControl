package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var errExpenseNil error = errors.New("expense is nil")

type ExpenseRepository interface {
	List(filter models.ExpenseFilter) ([]models.Expense, error)
	GetByID(id uint) (*models.Expense, error)
	Create(expense *models.Expense) error
	Update(expense *models.Expense) error
	Delete(id uint) error
	WithTx(tx TxProvider) ExpenseRepository
}

type gormExpenseRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewExpenseRepository(db *gorm.DB, logger *slog.Logger) ExpenseRepository {
	return &gormExpenseRepository{db: db, logger: logger}
}

func (r *gormExpenseRepository) WithTx(tx TxProvider) ExpenseRepository {
	return &gormExpenseRepository{db: tx.DB(), logger: r.logger}
}

func (r *gormExpenseRepository) List(filter models.ExpenseFilter) ([]models.Expense, error) {

	r.logger.Debug("repo.expense.list",
		slog.String("op", "repo.expense.list"),
	)

	var expenses []models.Expense
	query := r.db.Model(&models.Expense{}).Preload("Category").Where("user_id = ?", filter.UserID)

	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.StartDate != nil {
		query = query.Where("date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", *filter.EndDate)
	}
	if filter.MinAmount != nil {
		query = query.Where("amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		query = query.Where("amount <= ?", *filter.MaxAmount)
	}
	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	}
	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	}

	if err := query.Find(&expenses).Error; err != nil {
		r.logger.Error("repo.expense.list failed",
			slog.String("op", "repo.expense.list"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return expenses, nil
}

func (r *gormExpenseRepository) GetByID(id uint) (*models.Expense, error) {
	r.logger.Debug("repo.expense.get_by_id",
		slog.String("op", "repo.expense.get_by_id"),
		slog.Uint64("id", uint64(id)),
	)
	var expense models.Expense
	if err := r.db.First(&expense, id).Error; err != nil {
		r.logger.Error("repo.expense.get_by_id failed",
			slog.String("op", "repo.expense.get_by_id"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &expense, nil
}

func (r *gormExpenseRepository) Create(expense *models.Expense) error {
	if expense == nil {
		return errExpenseNil
	}

	r.logger.Debug("repo.expense.create",
		slog.String("op", "repo.expense.create"),
		slog.Uint64("user_id", uint64(expense.UserID)),
		slog.Float64("amount", expense.Amount),
		slog.Uint64("category", uint64(expense.CategoryID)),
	)

	if err := r.db.Create(expense).Error; err != nil {
		r.logger.Error("repo.expense.create failed",
			slog.String("op", "repo.expense.create"),
			slog.Uint64("user_id", uint64(expense.UserID)),
			slog.Float64("amount", expense.Amount),
			slog.Uint64("category", uint64(expense.CategoryID)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormExpenseRepository) Update(expense *models.Expense) error {
	if expense == nil {
		return errExpenseNil
	}
	r.logger.Debug("repo.expense.update",
		slog.String("op", "repo.expense.update"),
		slog.Uint64("id", uint64(expense.ID)),
	)

	if err := r.db.Save(expense).Error; err != nil {
		r.logger.Error("repo.expense.update failed",
			slog.String("op", "repo.expense.update"),
			slog.Uint64("id", uint64(expense.ID)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormExpenseRepository) Delete(id uint) error {
	r.logger.Debug("repo.expense.delete",
		slog.String("op", "repo.expense.delete"),
		slog.Uint64("id", uint64(id)),
	)
	if err := r.db.Delete(&models.Expense{}, id).Error; err != nil {
		r.logger.Error("repo.expense.delete failed",
			slog.String("op", "repo.expense.delete"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
