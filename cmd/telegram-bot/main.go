package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Command() {
		case "start":
			handleStart(bot, update.Message.Chat.ID)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Напиши /start 🙂")
			bot.Send(msg)
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(
		chatID,
		"👋 Привет!\n\nНажми кнопку *Menu* внизу, чтобы открыть CashControl 💸",
	)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
