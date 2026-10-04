package bot

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// FakeCompleter records the last prompt and returns a canned answer.
type FakeCompleter struct {
	LastPrompt string
	Answer     string
	Err        error
}

func (f *FakeCompleter) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	f.LastPrompt = prompt
	return f.Answer, f.Err
}

func TestValidateContact(t *testing.T) {
	if err := ValidateContact(""); err == nil {
		t.Error("Expected error for empty contact")
	}
	if err := ValidateContact("organizer@example.com"); err == nil {
		t.Error("Expected error for contact containing example.com")
	}
	if err := ValidateContact("real@myuniversity.edu"); err != nil {
		t.Error("Expected valid contact to pass")
	}
}

func TestHandler_AllowRateLimit(t *testing.T) {
	h := NewHandler(nil, nil, "", "test@myuniversity.edu")
	
	if !h.allow(123) {
		t.Error("First request should be allowed")
	}
	
	if h.allow(123) {
		t.Error("Second request within 3 seconds should be blocked")
	}
	
	if !h.allow(456) {
		t.Error("First request from different user should be allowed")
	}
}

// MockTransport intercepts HTTP requests
type MockTransport struct {
	sentMessages *[]string
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "getMe") {
		jsonBody := `{"ok": true, "result": {"id": 12345, "is_bot": true, "first_name": "TestBot", "username": "test_bot"}}`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(jsonBody)),
		}, nil
	}

	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		bodyStr, _ := url.QueryUnescape(string(bodyBytes))
		*m.sentMessages = append(*m.sentMessages, bodyStr)
	}
	// Return a dummy valid JSON response
	jsonBody := `{"ok": true, "result": {"message_id": 1}}`
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(jsonBody)),
	}, nil
}

func TestHandler_MarkersAndRates(t *testing.T) {
	var sentMessages []string
	
	client := &http.Client{
		Transport: &MockTransport{sentMessages: &sentMessages},
	}

	botAPI, err := tgbotapi.NewBotAPIWithClient("dummy", tgbotapi.APIEndpoint, client)
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	comp := &FakeCompleter{Answer: "NOINFO|EN"}
	h := NewHandler(botAPI, comp, "../../event.md", "real@myuniversity.edu")

	// Create dummy update with text
	makeUpdate := func(text string, isCmd bool) tgbotapi.Update {
		var entities []tgbotapi.MessageEntity
		if isCmd {
			entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(text)}}
		}
		return tgbotapi.Update{
			Message: &tgbotapi.Message{
				MessageID: 1,
				From:      &tgbotapi.User{ID: 123},
				Chat:      &tgbotapi.Chat{ID: 456},
				Text:      text,
				Entities:  entities,
			},
		}
	}

	// 1. NOINFO|EN mapping
	h.HandleUpdate(makeUpdate("Test question", false))
	if len(sentMessages) == 0 || !strings.Contains(sentMessages[0], "I don't have that info") {
		t.Errorf("Expected English fallback, got: %v", sentMessages)
	}

	// 2. Command bypasses rate limiter
	// First message was sent, so normal message would be blocked
	sentMessages = nil
	h.HandleUpdate(makeUpdate("/help", true))
	if len(sentMessages) == 0 || !strings.Contains(sentMessages[0], "Commands:") {
		t.Errorf("Expected command bypass and help msg, got: %v", sentMessages)
	}

	// 3. Normal text blocked due to rate limit, sends warning
	sentMessages = nil
	h.HandleUpdate(makeUpdate("Another test", false))
	if len(sentMessages) == 0 || !strings.Contains(sentMessages[0], "wait a few seconds") {
		t.Errorf("Expected rate limit warning, got: %v", sentMessages)
	}
	
	// 4. Rate limit warning not sent again immediately
	sentMessages = nil
	h.HandleUpdate(makeUpdate("Yet another test", false))
	if len(sentMessages) > 0 {
		t.Errorf("Expected no message (rate limit warning throttled), got: %v", sentMessages)
	}

	// 5. Non-text update
	sentMessages = nil
	nonTextUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 999}, // different user
			Chat: &tgbotapi.Chat{ID: 456},
			Photo: []tgbotapi.PhotoSize{{FileID: "123"}},
		},
	}
	h.HandleUpdate(nonTextUpdate)
	if len(sentMessages) == 0 || !strings.Contains(sentMessages[0], "I can only read text messages") {
		t.Errorf("Expected text-only warning, got: %v", sentMessages)
	}
}
