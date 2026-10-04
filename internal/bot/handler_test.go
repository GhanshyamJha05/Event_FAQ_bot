package bot

import (
	"context"
	"testing"
	"time"
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

func TestHandler_AllowRateLimit(t *testing.T) {
	h := NewHandler(nil, nil, "", "")
	
	// First request should be allowed
	if !h.allow(123) {
		t.Error("First request should be allowed")
	}
	
	// Immediate second request should be blocked
	if h.allow(123) {
		t.Error("Second request within 3 seconds should be blocked")
	}
	
	// Different user should be allowed
	if !h.allow(456) {
		t.Error("First request from different user should be allowed")
	}

	// Manually advance time by modifying last seen to simulate time passing
	h.mu.Lock()
	h.userLastSeen[123] = time.Now().Add(-4 * time.Second)
	h.mu.Unlock()

	// Third request after 4 seconds should be allowed
	if !h.allow(123) {
		t.Error("Request after 3 seconds should be allowed")
	}
}

func TestHandler_PromptConstruction(t *testing.T) {
	// Not full end-to-end because tgbotapi types are tricky to instantiate fully 
	// for routing, but we can test the structure logically or just rely on the 
	// real model checks for prompt correctness.
	
	h := NewHandler(nil, nil, "event.md", "org@example.com")
	
	// Checking the fallback msg creation
	expectedFallback := "I don't have that info. Please ask the organizer: org@example.com"
	if h.FallbackMsg != expectedFallback {
		t.Errorf("Expected fallback %q, got %q", expectedFallback, h.FallbackMsg)
	}
}
