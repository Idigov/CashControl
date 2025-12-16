package services

import (
	"cashcontrol/internal/repository"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type NotificationService interface {
	SendToChat(chatID int64, text string) error
	SendToUser(userID uint, text string) error
}

type telegramNotificationService struct {
	bot    *tgbotapi.BotAPI
	users  repository.UserRepository
	logger *slog.Logger
}

func NewNotificationService(botToken string, users repository.UserRepository, logger *slog.Logger) (NotificationService, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	return &telegramNotificationService{
		bot:    bot,
		users:  users,
		logger: logger,
	}, nil
}

func (s *telegramNotificationService) SendToChat(chatID int64, text string) error {
	if chatID == 0 {
		return fmt.Errorf("chat id is empty")
	}
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := s.bot.Send(msg)
	if err != nil {
		s.logger.Warn("send telegram message failed", slog.Int64("chat_id", chatID), slog.String("error", err.Error()))
	}
	return err
}

func (s *telegramNotificationService) SendToUser(userID uint, text string) error {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}
	if user.TelegramChatID != nil {
		return s.SendToChat(*user.TelegramChatID, text)
	}
	if user.TelegramID != nil {
		return s.SendToChat(*user.TelegramID, text)
	}
	return fmt.Errorf("no chat id for user %d", userID)
}

