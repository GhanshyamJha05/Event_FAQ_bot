package main

import (
	"log"
	"os"

	"event-faq-bot/internal/bot"
	"event-faq-bot/internal/llm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	llmBaseURL := os.Getenv("LLM_BASE_URL")
	if llmBaseURL == "" {
		llmBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
	}

	llmAPIKey := os.Getenv("LLM_API_KEY")
	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		log.Fatal("LLM_MODEL is required")
	}

	organizerContact := os.Getenv("ORGANIZER_CONTACT")
	if err := bot.ValidateContact(organizerContact); err != nil {
		log.Fatalf("Invalid contact: %v", err)
	}

	eventFile := os.Getenv("EVENT_FILE")
	if eventFile == "" {
		eventFile = "event.md"
	}

	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", tgBot.Self.UserName)

	completer := llm.NewClient(llmBaseURL, llmAPIKey, llmModel)
	handler := bot.NewHandler(tgBot, completer, eventFile, organizerContact)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tgBot.GetUpdatesChan(u)

	for update := range updates {
		go handler.HandleUpdate(update)
	}
}
