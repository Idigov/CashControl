package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("неверные учетные данные")

type AuthService interface {
	Register(req models.RegisterRequest) (*models.LoginResponse, error)
	Login(req models.LoginRequest) (*models.LoginResponse, error)
}

type authService struct {
	users     repository.UserRepository
	logger    *slog.Logger
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, logger *slog.Logger, jwtSecret string) AuthService {
	return &authService{users: users, logger: logger, jwtSecret: jwtSecret}
}

func (s *authService) Register(req models.RegisterRequest) (*models.LoginResponse, error) {
	// простая валидация
	if err := s.validateRegister(req); err != nil {
		return nil, err
	}

	// проверим, что пользователь с email не существует
	u, err := s.users.GetByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("failed to check email existence",
			slog.String("email", req.Email),
			slog.String("error", err.Error()),
		)
		return nil, errors.New("ошибка при проверке email")
	}
	if u != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// хешируем пароль
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", slog.String("error", err.Error()))
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashed),
	}

	if err := s.users.Create(user); err != nil {
		s.logger.Error("failed to create user", slog.String("error", err.Error()))
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{Token: token, User: user}, nil
}

func (s *authService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.users.GetByEmail(req.Email)
	if err != nil {
		s.logger.Warn("user not found during login", slog.String("email", req.Email))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("invalid password for user", slog.Uint64("user_id", uint64(user.ID)))
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{Token: token, User: user}, nil
}

func (s *authService) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("failed to sign token", slog.String("error", err.Error()))
		return "", err
	}
	return signed, nil
}

func (s *authService) validateRegister(req models.RegisterRequest) error {
	if req.Email == "" {
		return errors.New("email не может быть пустым")
	}
	if req.Username == "" {
		return errors.New("username не может быть пустым")
	}
	if len(req.Password) < 6 {
		return errors.New("пароль должен быть не менее 6 символов")
	}
	return nil
}
