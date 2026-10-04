package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"event-faq-bot/internal/eventinfo"
	"event-faq-bot/internal/llm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler manages the bot logic.
type Handler struct {
	Bot              *tgbotapi.BotAPI
	Completer        llm.Completer
	EventFile        string
	OrganizerContact string
	FallbackMsg      string

	// Rate limiting state
	userLastSeen map[int64]time.Time
	mu           sync.Mutex
}

// NewHandler creates a new bot handler.
func NewHandler(bot *tgbotapi.BotAPI, completer llm.Completer, eventFile, organizerContact string) *Handler {
	fallback := fmt.Sprintf("I don't have that info. Please ask the organizer: %s", organizerContact)
	return &Handler{
		Bot:              bot,
		Completer:        completer,
		EventFile:        eventFile,
		OrganizerContact: organizerContact,
		FallbackMsg:      fallback,
		userLastSeen:     make(map[int64]time.Time),
	}
}

// HandleUpdate processes an incoming Telegram update.
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	// Check rate limit
	if !h.allow(userID) {
		// Ignore if rate limited
		return
	}

	if update.Message.IsCommand() {
		h.handleCommand(update.Message)
		return
	}

	if update.Message.Text == "" {
		h.reply(chatID, update.Message.MessageID, "Please send a text message or a question.")
		return
	}

	// Handle standard question
	h.handleQuestion(update.Message)
}

func (h *Handler) allow(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	lastSeen, exists := h.userLastSeen[userID]
	now := time.Now()
	if exists && now.Sub(lastSeen) < 3*time.Second {
		return false
	}
	h.userLastSeen[userID] = now
	return true
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	switch msg.Command() {
	case "start":
		h.reply(chatID, msg.MessageID, "Hello! I am the event FAQ bot. Ask me any question about the event and I'll do my best to answer. Use /help to see more commands.")
	case "help":
		h.reply(chatID, msg.MessageID, "Commands:\n/start - Greeting\n/help - Show commands\n/details - Show full event details\nOr just type your question!")
	case "details":
		info, err := eventinfo.Load(h.EventFile)
		if err != nil {
			h.reply(chatID, msg.MessageID, "Sorry, I couldn't load the event details right now.")
			return
		}
		h.reply(chatID, msg.MessageID, info)
	default:
		h.reply(chatID, msg.MessageID, "I don't know that command.")
	}
}

func (h *Handler) handleQuestion(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	question := msg.Text

	// Load event info on every question
	info, err := eventinfo.Load(h.EventFile)
	if err != nil {
		log.Printf("Failed to load event info: %v", err)
		h.reply(chatID, msg.MessageID, "I'm having trouble reading the event details right now. Please try again later.")
		return
	}

	prompt := fmt.Sprintf(`You are the assistant for the event described below.
Answer ONLY using the event text. 
If the answer is not clearly stated or you don't know, reply EXACTLY with this fallback message: "%s"
Keep answers under 4 sentences.
Reply in the same language/style the user wrote in (English or Hinglish).
NEVER invent times, places, prices, or rules.

Event details:
%s

Question: %s`, h.FallbackMsg, info, question)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	answer, err := h.Completer.ChatCompletion(ctx, prompt)
	if err != nil {
		log.Printf("LLM completion error: %v", err)
		failMsg := fmt.Sprintf("I'm having trouble right now, please ask the organizer: %s", h.OrganizerContact)
		h.reply(chatID, msg.MessageID, failMsg)
		return
	}

	h.reply(chatID, msg.MessageID, answer)
	h.logInteraction(question, answer)
}

func (h *Handler) reply(chatID int64, replyToID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToID
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Fallback  bool   `json:"fallback"`
}

func (h *Handler) logInteraction(question, answer string) {
	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Question:  question,
		Answer:    answer,
		Fallback:  answer == h.FallbackMsg,
	}
	
	b, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	f, err := os.OpenFile("questions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(append(b, '\n')); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}
