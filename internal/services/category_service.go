package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var ErrCategoryNotFound = errors.New("категория не найдена")

type CategoryService interface {
	CreateCategory(userID int, req models.CreateCategoryRequest) (*models.Category, error)
	GetCategoryList(userID int) ([]models.Category, error)
	GetCategoryByID(id uint) (*models.Category, error)
	UpdateCategory(id uint, req models.UpdateCategoryRequest) (*models.Category, error)
	DeleteCategory(id uint) error
}

type categoryService struct {
	categories repository.CategoryRepository
	logger     *slog.Logger
}

func NewCategoryService(categories repository.CategoryRepository, logger *slog.Logger) CategoryService {
	return &categoryService{categories: categories, logger: logger}
}

func (s *categoryService) CreateCategory(userID int, req models.CreateCategoryRequest) (*models.Category, error) {
	if err := s.validateCategoryCreate(req); err != nil {
		s.logger.Warn("category create validation failed",
			slog.Int("user_id", userID),
			slog.String("name", req.Name),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	category := &models.Category{
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
		Icon:   req.Icon,
	}

	if err := s.categories.Create(category); err != nil {
		s.logger.Error("category create failed",
			slog.String("op", "create_category"),
			slog.Int("user_id", userID),
			slog.String("name", req.Name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("category created",
		slog.Uint64("category_id", uint64(category.ID)),
		slog.Int("user_id", userID),
		slog.String("name", category.Name),
	)

	return category, nil
}

func (s *categoryService) GetCategoryList(userID int) ([]models.Category, error) {
	categories, err := s.categories.GetByUserID(userID)
	if err != nil {
		s.logger.Error("failed to list categories",
			slog.String("op", "list_categories"),
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("categories listed",
		slog.Int("user_id", userID),
		slog.Int("count", len(categories)),
	)

	return categories, nil
}

func (s *categoryService) GetCategoryByID(id uint) (*models.Category, error) {
	category, err := s.categories.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("category not found",
				slog.Uint64("category_id", uint64(id)),
			)
			return nil, ErrCategoryNotFound
		}
		s.logger.Error("failed to get category",
			slog.String("op", "get_category_by_id"),
			slog.Uint64("category_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("category retrieved",
		slog.Uint64("category_id", uint64(id)),
		slog.String("name", category.Name),
	)

	return category, nil
}

func (s *categoryService) UpdateCategory(id uint, req models.UpdateCategoryRequest) (*models.Category, error) {
	category, err := s.categories.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("category not found for update",
				slog.Uint64("category_id", uint64(id)),
			)
			return nil, ErrCategoryNotFound
		}
		s.logger.Error("failed to fetch category before update",
			slog.String("op", "update_category"),
			slog.Uint64("category_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Color != nil {
		category.Color = *req.Color
	}
	if req.Icon != nil {
		category.Icon = *req.Icon
	}

	if err := s.categories.Update(category); err != nil {
		s.logger.Error("category update failed",
			slog.String("op", "update_category"),
			slog.Uint64("category_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("category updated",
		slog.Uint64("category_id", uint64(id)),
		slog.String("name", category.Name),
	)

	return category, nil
}

func (s *categoryService) DeleteCategory(id uint) error {
	_, err := s.categories.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("category not found for delete",
				slog.Uint64("category_id", uint64(id)),
			)
			return ErrCategoryNotFound
		}
		s.logger.Error("failed to fetch category before delete",
			slog.String("op", "delete_category"),
			slog.Uint64("category_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.categories.Delete(id); err != nil {
		s.logger.Error("category delete failed",
			slog.String("op", "delete_category"),
			slog.Uint64("category_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("category deleted",
		slog.Uint64("category_id", uint64(id)),
	)

	return nil
}

func (s *categoryService) validateCategoryCreate(req models.CreateCategoryRequest) error {
	if req.Name == "" {
		return errors.New("название категории не может быть пустым")
	}
	return nil
}
