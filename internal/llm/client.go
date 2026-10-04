package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Completer defines the interface for an LLM client.
type Completer interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

// Client implements the Completer interface for OpenAI-compatible APIs.
type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewClient creates a new OpenAI-compatible LLM client.
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// chatRequest represents the JSON payload for the chat completion request.
type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse represents the JSON payload of the chat completion response.
type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

// errorResponse represents an API error response.
type errorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion sends a prompt to the LLM and returns the generated text.
// It implements a single retry on 429/5xx errors with a short backoff.
func (c *Client) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	// Bundle everything into a single user message to avoid system/developer role restrictions.
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   1024,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"

	var lastErr error
	for attempt := 1; attempt <= 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if c.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.APIKey)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read response body: %w", err)
			}

			var cr chatResponse
			if err := json.Unmarshal(body, &cr); err != nil {
				return "", fmt.Errorf("failed to parse response: %w", err)
			}

			if len(cr.Choices) > 0 {
				content := cr.Choices[0].Message.Content
				
				// Strip <thought>...</thought> blocks
				for {
					start := strings.Index(content, "<thought>")
					if start == -1 {
						break
					}
					
					end := strings.Index(content[start:], "</thought>")
					if end == -1 {
						// Unclosed <thought> block found
						return "", fmt.Errorf("unclosed <thought> block in response")
					}
					
					// Reconstruct string without the thought block
					content = content[:start] + content[start+end+10:]
				}
				
				return strings.TrimSpace(content), nil
			}
			return "", fmt.Errorf("no completion choices returned")
		}

		// Handle errors
		body, _ := io.ReadAll(resp.Body)
		var apiErr errorResponse
		_ = json.Unmarshal(body, &apiErr)
		
		errMsg := apiErr.Error.Message
		if errMsg == "" {
			errMsg = string(body)
		}

		lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, errMsg)

		// Retry only on 429 or 5xx errors
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt < 2 {
				time.Sleep(2 * time.Second) // short backoff
				continue
			}
		} else {
			// Don't retry on 4xx errors other than 429
			break
		}
	}

	return "", lastErr
}
