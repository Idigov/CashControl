package services

import (
	"cashcontrol/internal/models"
	"cashcontrol/internal/repository"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("неверные учетные данные")

type AuthService interface {
	Register(req models.RegisterRequest) (*models.LoginResponse, error)
	Login(req models.LoginRequest) (*models.LoginResponse, error)
	LoginWithTelegram(initData string) (*models.LoginResponse, error)
}

type authService struct {
	users            repository.UserRepository
	categories       repository.CategoryRepository
	logger           *slog.Logger
	jwtSecret        string
	telegramBotToken string
}

func NewAuthService(
	users repository.UserRepository,
	categories repository.CategoryRepository,
	logger *slog.Logger,
	jwtSecret string,
	telegramBotToken string,
) AuthService {
	return &authService{
		users:            users,
		categories:       categories,
		logger:           logger,
		jwtSecret:        jwtSecret,
		telegramBotToken: telegramBotToken,
	}
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

	email := req.Email
	username := req.Username
	password := string(hashed)

	user := &models.User{
		Email:    &email,
		Username: &username,
		Password: &password,
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
	if user.Password == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
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

func (s *authService) LoginWithTelegram(initData string) (*models.LoginResponse, error) {
	telegramID, err := validateTelegramInitData(initData, s.telegramBotToken)
	if err != nil {
		s.logger.Warn("telegram auth failed", slog.String("error", err.Error()))
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByTelegramID(telegramID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		user = &models.User{
			TelegramID: &telegramID,
			TelegramChatID: &telegramID,
		}
		if err := s.users.Create(user); err != nil {
			return nil, err
		}

		s.createDefaultCategories(user.ID)
	} else {
		// Проставляем chat_id, если его ещё нет
		if user.TelegramChatID == nil {
			user.TelegramChatID = &telegramID
			_ = s.users.Update(user)
		}
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func validateTelegramInitData(initData, botToken string) (int64, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, err
	}

	hash := values.Get("hash")
	if hash == "" {
		return 0, errors.New("hash missing")
	}
	values.Del("hash")

	// ⏱ Проверка auth_date
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return 0, errors.New("auth_date missing")
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid auth_date")
	}

	if time.Now().Unix()-authDate > 86400 {
		return 0, errors.New("telegram auth expired")
	}

	// 📦 Формируем data_check_string
	var pairs []string
	for k, v := range values {
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)

	dataCheckString := strings.Join(pairs, "\n")

	// 🔑 КЛЮЧЕВОЕ ОТЛИЧИЕ ДЛЯ MINI APP
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secret := secretKey.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))

	calculatedHash := hex.EncodeToString(h.Sum(nil))

	if calculatedHash != hash {
		return 0, errors.New("invalid telegram signature")
	}

	// 👤 Извлекаем Telegram ID
	if idStr := values.Get("id"); idStr != "" {
		return strconv.ParseInt(idStr, 10, 64)
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return 0, errors.New("telegram user missing")
	}

	var tgUser struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return 0, err
	}

	return tgUser.ID, nil
}

func (s *authService) createDefaultCategories(userID uint) {
	defaults := []models.Category{
		{Name: "Еда", Color: "#F97316", Icon: "🍔"},
		{Name: "Транспорт", Color: "#0EA5E9", Icon: "🚕"},
		{Name: "Дом", Color: "#22C55E", Icon: "🏠"},
		{Name: "Подписки", Color: "#8B5CF6", Icon: "📱"},
	}

	for _, c := range defaults {
		category := c
		category.UserID = userID
		category.IsDefault = true
		_ = s.categories.Create(&category)
	}
}
