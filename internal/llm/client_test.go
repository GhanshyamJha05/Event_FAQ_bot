package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ChatCompletion_ThoughtStripping(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		expectErr  bool
		expectText string
	}{
		{
			name:       "normal case, no thought block",
			response:   "This is a normal answer.",
			expectErr:  false,
			expectText: "This is a normal answer.",
		},
		{
			name:       "single thought block",
			response:   "<thought>thinking...</thought>Here is the answer.",
			expectErr:  false,
			expectText: "Here is the answer.",
		},
		{
			name:       "multiple thought blocks",
			response:   "<thought>first</thought>Hello <thought>second</thought>World",
			expectErr:  false,
			expectText: "Hello World",
		},
		{
			name:       "unclosed thought block",
			response:   "<thought>thinking... but not closing it",
			expectErr:  true,
			expectText: "",
		},
		{
			name:       "thought block with newlines",
			response:   "<thought>\nline 1\nline 2\n</thought>\nFinal answer.",
			expectErr:  false,
			expectText: "Final answer.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				responseJSON := `{
					"choices": [{
						"message": {
							"content": "` + strings.ReplaceAll(strings.ReplaceAll(tt.response, "\n", "\\n"), "\"", "\\\"") + `"
						}
					}]
				}`
				w.Write([]byte(responseJSON))
			}))
			defer ts.Close()

			client := NewClient(ts.URL, "dummy", "dummy-model")
			ans, err := client.ChatCompletion(context.Background(), "test prompt")

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr && ans != tt.expectText {
				t.Errorf("expected text %q, got %q", tt.expectText, ans)
			}
		})
	}
}
