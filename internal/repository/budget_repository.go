package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var errBudgetNil error = errors.New("budget is nil")

type BudgetRepository interface {
	List() ([]models.Budget, error)
	GetByID(id uint) (*models.Budget, error)
	GetByUserIDAndMonth(userID int, month, year int) (*models.Budget, error)
	GetByUserID(userID int) ([]models.Budget, error)
	Create(budget *models.Budget) error
	Update(budget *models.Budget) error
	Delete(id uint) error
}

type gormBudgetRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewBudgetRepository(db *gorm.DB, logger *slog.Logger) BudgetRepository {
	return &gormBudgetRepository{db: db, logger: logger}
}

func (r *gormBudgetRepository) List() ([]models.Budget, error) {
	r.logger.Debug("repo.budget.list",
		slog.String("op", "repo.budget.list"),
	)
	var budgets []models.Budget
	if err := r.db.Find(&budgets).Error; err != nil {
		r.logger.Error("repo.budget.list failed",
			slog.String("op", "repo.budget.list"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return budgets, nil
}

func (r *gormBudgetRepository) GetByID(id uint) (*models.Budget, error) {
	r.logger.Debug("repo.budget.get_by_id",
		slog.String("op", "repo.budget.get_by_id"),
		slog.Uint64("id", uint64(id)),
	)
	var budget models.Budget
	if err := r.db.First(&budget, id).Error; err != nil {
		r.logger.Error("repo.budget.get_by_id failed",
			slog.String("op", "repo.budget.get_by_id"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &budget, nil
}

func (r *gormBudgetRepository) GetByUserIDAndMonth(userID int, month, year int) (*models.Budget, error) {
	r.logger.Debug("repo.budget.get_by_user_id_and_month",
		slog.String("op", "repo.budget.get_by_user_id_and_month"),
		slog.Int("user_id", userID),
		slog.Int("month", month),
		slog.Int("year", year),
	)
	var budget models.Budget
	if err := r.db.Where("user_id = ? AND month = ? AND year = ?", userID, month, year).First(&budget).Error; err != nil {
		r.logger.Error("repo.budget.get_by_user_id_and_month failed",
			slog.String("op", "repo.budget.get_by_user_id_and_month"),
			slog.Int("user_id", userID),
			slog.Int("month", month),
			slog.Int("year", year),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &budget, nil
}

func (r *gormBudgetRepository) GetByUserID(userID int) ([]models.Budget, error) {
	r.logger.Debug("repo.budget.get_by_user_id",
		slog.String("op", "repo.budget.get_by_user_id"),
		slog.Int("user_id", userID),
	)
	var budgets []models.Budget
	if err := r.db.Where("user_id = ?", userID).Order("year DESC, month DESC").Find(&budgets).Error; err != nil {
		r.logger.Error("repo.budget.get_by_user_id failed",
			slog.String("op", "repo.budget.get_by_user_id"),
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return budgets, nil
}

func (r *gormBudgetRepository) Create(budget *models.Budget) error {
	if budget == nil {
		return errBudgetNil
	}

	r.logger.Debug("repo.budget.create",
		slog.String("op", "repo.budget.create"),
		slog.Int("user_id", budget.UserID),
		slog.Float64("amount", budget.Amount),
		slog.Int("month", budget.Month),
		slog.Int("year", budget.Year),
	)

	if err := r.db.Create(budget).Error; err != nil {
		r.logger.Error("repo.budget.create failed",
			slog.String("op", "repo.budget.create"),
			slog.Int("user_id", budget.UserID),
			slog.Float64("amount", budget.Amount),
			slog.Int("month", budget.Month),
			slog.Int("year", budget.Year),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormBudgetRepository) Update(budget *models.Budget) error {
	if budget == nil {
		return errBudgetNil
	}
	r.logger.Debug("repo.budget.update",
		slog.String("op", "repo.budget.update"),
		slog.Uint64("id", uint64(budget.ID)),
	)

	if err := r.db.Save(budget).Error; err != nil {
		r.logger.Error("repo.budget.update failed",
			slog.String("op", "repo.budget.update"),
			slog.Uint64("id", uint64(budget.ID)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormBudgetRepository) Delete(id uint) error {
	r.logger.Debug("repo.budget.delete",
		slog.String("op", "repo.budget.delete"),
		slog.Uint64("id", uint64(id)),
	)
	if err := r.db.Delete(&models.Budget{}, id).Error; err != nil {
		r.logger.Error("repo.budget.delete failed",
			slog.String("op", "repo.budget.delete"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
