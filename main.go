package main

import (
	"log"
	"os"

	"pomodoromebot/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Printf("Authorized on account: @%s", api.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := api.GetUpdatesChan(updateConfig)

	for update := range updates {
		bot.HandleUpdate(api, update)
	}
}
