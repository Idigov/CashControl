package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var errCategoryNil error = errors.New("category is nil")

type CategoryRepository interface {
	List() ([]models.Category, error)
	GetByID(id uint) (*models.Category, error)
	GetByUserID(userID int) ([]models.Category, error)
	Create(category *models.Category) error
	Update(category *models.Category) error
	Delete(id uint) error
}

type gormCategoryRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewCategoryRepository(db *gorm.DB, logger *slog.Logger) CategoryRepository {
	return &gormCategoryRepository{db: db, logger: logger}
}

func (r *gormCategoryRepository) List() ([]models.Category, error) {
	r.logger.Debug("repo.category.list",
		slog.String("op", "repo.category.list"),
	)
	var categories []models.Category
	if err := r.db.Find(&categories).Error; err != nil {
		r.logger.Error("repo.category.list failed",
			slog.String("op", "repo.category.list"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return categories, nil
}

func (r *gormCategoryRepository) GetByID(id uint) (*models.Category, error) {
	r.logger.Debug("repo.category.get_by_id",
		slog.String("op", "repo.category.get_by_id"),
		slog.Uint64("id", uint64(id)),
	)
	var category models.Category
	if err := r.db.First(&category, id).Error; err != nil {
		r.logger.Error("repo.category.get_by_id failed",
			slog.String("op", "repo.category.get_by_id"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &category, nil
}

func (r *gormCategoryRepository) GetByUserID(userID int) ([]models.Category, error) {
	r.logger.Debug("repo.category.get_by_user_id",
		slog.String("op", "repo.category.get_by_user_id"),
		slog.Int("user_id", userID),
	)
	var categories []models.Category
	if err := r.db.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		r.logger.Error("repo.category.get_by_user_id failed",
			slog.String("op", "repo.category.get_by_user_id"),
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return categories, nil
}

func (r *gormCategoryRepository) Create(category *models.Category) error {
	if category == nil {
		return errCategoryNil
	}

	r.logger.Debug("repo.category.create",
		slog.String("op", "repo.category.create"),
		slog.Uint64("user_id", uint64(category.UserID)),
		slog.String("name", category.Name),
	)

	if err := r.db.Create(category).Error; err != nil {
		r.logger.Error("repo.category.create failed",
			slog.String("op", "repo.category.create"),
			slog.Uint64("user_id", uint64(category.UserID)),
			slog.String("name", category.Name),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormCategoryRepository) Update(category *models.Category) error {
	if category == nil {
		return errCategoryNil
	}
	r.logger.Debug("repo.category.update",
		slog.String("op", "repo.category.update"),
		slog.Uint64("id", uint64(category.ID)),
	)

	if err := r.db.Save(category).Error; err != nil {
		r.logger.Error("repo.category.update failed",
			slog.String("op", "repo.category.update"),
			slog.Uint64("id", uint64(category.ID)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormCategoryRepository) Delete(id uint) error {
	r.logger.Debug("repo.category.delete",
		slog.String("op", "repo.category.delete"),
		slog.Uint64("id", uint64(id)),
	)
	if err := r.db.Delete(&models.Category{}, id).Error; err != nil {
		r.logger.Error("repo.category.delete failed",
			slog.String("op", "repo.category.delete"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
