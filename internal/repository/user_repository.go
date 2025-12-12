package repository

import (
	"cashcontrol/internal/models"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var errUserNil error = errors.New("user is nil")

type UserRepository interface {
	List() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uint) error
}

type gormUserRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewUserRepository(db *gorm.DB, logger *slog.Logger) UserRepository {
	return &gormUserRepository{db: db, logger: logger}
}

func (r *gormUserRepository) List() ([]models.User, error) {
	r.logger.Debug("repo.user.list",
		slog.String("op", "repo.user.list"),
	)
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		r.logger.Error("repo.user.list failed",
			slog.String("op", "repo.user.list"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) GetByID(id uint) (*models.User, error) {
	r.logger.Debug("repo.user.get_by_id",
		slog.String("op", "repo.user.get_by_id"),
		slog.Uint64("id", uint64(id)),
	)
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		r.logger.Error("repo.user.get_by_id failed",
			slog.String("op", "repo.user.get_by_id"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) GetByEmail(email string) (*models.User, error) {
	r.logger.Debug("repo.user.get_by_email",
		slog.String("op", "repo.user.get_by_email"),
		slog.String("email", email),
	)
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		r.logger.Error("repo.user.get_by_email failed",
			slog.String("op", "repo.user.get_by_email"),
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Create(user *models.User) error {
	if user == nil {
		return errUserNil
	}

	r.logger.Debug("repo.user.create",
		slog.String("op", "repo.user.create"),
		slog.String("email", user.Email),
		slog.String("username", user.Username),
	)

	if err := r.db.Create(user).Error; err != nil {
		r.logger.Error("repo.user.create failed",
			slog.String("op", "repo.user.create"),
			slog.String("email", user.Email),
			slog.String("username", user.Username),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormUserRepository) Update(user *models.User) error {
	if user == nil {
		return errUserNil
	}
	r.logger.Debug("repo.user.update",
		slog.String("op", "repo.user.update"),
		slog.Uint64("id", uint64(user.ID)),
	)

	if err := r.db.Save(user).Error; err != nil {
		r.logger.Error("repo.user.update failed",
			slog.String("op", "repo.user.update"),
			slog.Uint64("id", uint64(user.ID)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (r *gormUserRepository) Delete(id uint) error {
	r.logger.Debug("repo.user.delete",
		slog.String("op", "repo.user.delete"),
		slog.Uint64("id", uint64(id)),
	)
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		r.logger.Error("repo.user.delete failed",
			slog.String("op", "repo.user.delete"),
			slog.Uint64("id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
