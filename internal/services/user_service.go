package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("пользователь не найден")

type UserService interface {
	CreateUser(req models.RegisterRequest) (*models.User, error)
	GetUserList() ([]models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, email, username string) (*models.User, error)
	DeleteUser(id uint) error
}

type userService struct {
	users  repository.UserRepository
	logger *slog.Logger
}

func NewUserService(users repository.UserRepository, logger *slog.Logger) UserService {
	return &userService{users: users, logger: logger}
}

func (s *userService) CreateUser(req models.RegisterRequest) (*models.User, error) {
	if err := s.validateUserCreate(req); err != nil {
		s.logger.Warn("user create validation failed",
			slog.String("email", req.Email),
			slog.String("username", req.Username),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password, // В реальном приложении должно быть хешировано
	}

	if err := s.users.Create(user); err != nil {
		s.logger.Error("user create failed",
			slog.String("op", "create_user"),
			slog.String("email", req.Email),
			slog.String("username", req.Username),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("user created",
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("email", user.Email),
		slog.String("username", user.Username),
	)

	return user, nil
}

func (s *userService) GetUserList() ([]models.User, error) {
	users, err := s.users.List()
	if err != nil {
		s.logger.Error("failed to list users",
			slog.String("op", "list_users"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("users listed",
		slog.Int("count", len(users)),
	)

	return users, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("user not found",
				slog.Uint64("user_id", uint64(id)),
			)
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user",
			slog.String("op", "get_user_by_id"),
			slog.Uint64("user_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("user retrieved",
		slog.Uint64("user_id", uint64(id)),
		slog.String("email", user.Email),
		slog.String("username", user.Username),
	)

	return user, nil
}

func (s *userService) UpdateUser(id uint, email, username string) (*models.User, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("user not found for update",
				slog.Uint64("user_id", uint64(id)),
			)
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to fetch user before update",
			slog.String("op", "update_user"),
			slog.Uint64("user_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	if email != "" {
		user.Email = email
	}
	if username != "" {
		user.Username = username
	}

	if err := s.users.Update(user); err != nil {
		s.logger.Error("user update failed",
			slog.String("op", "update_user"),
			slog.Uint64("user_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.Info("user updated",
		slog.Uint64("user_id", uint64(id)),
		slog.String("email", user.Email),
		slog.String("username", user.Username),
	)

	return user, nil
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.users.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("user not found for delete",
				slog.Uint64("user_id", uint64(id)),
			)
			return ErrUserNotFound
		}
		s.logger.Error("failed to fetch user before delete",
			slog.String("op", "delete_user"),
			slog.Uint64("user_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	if err := s.users.Delete(id); err != nil {
		s.logger.Error("user delete failed",
			slog.String("op", "delete_user"),
			slog.Uint64("user_id", uint64(id)),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.Info("user deleted",
		slog.Uint64("user_id", uint64(id)),
	)

	return nil
}

func (s *userService) validateUserCreate(req models.RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email не может быть пустым")
	}
	if req.Username == "" {
		return errors.New("username не может быть пустым")
	}
	if req.Password == "" {
		return errors.New("password не может быть пустым")
	}
	if len(req.Password) < 6 {
		return errors.New("пароль должен быть не менее 6 символов")
	}
	return nil
}
