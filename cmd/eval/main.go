package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"event-faq-bot/internal/eventinfo"
	"event-faq-bot/internal/llm"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		log.Fatal("LLM_API_KEY not set")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	client := llm.NewClient("https://generativelanguage.googleapis.com/v1beta/openai", apiKey, model)
	info, err := eventinfo.Load("event.md")
	if err != nil {
		log.Fatalf("Failed to load event.md: %v", err)
	}

	fallbackMsg := "I don't have that info. Please ask the organizer: organizer@example.com"

	questions := []struct {
		Q           string
		Answerable  bool
	}{
		// Answerable (English)
		{"When is the hackathon?", true},
		{"Where is the venue?", true},
		{"What is the team size limit?", true},
		{"Do I need to bring my laptop?", true},
		// Answerable (Hinglish)
		{"Bhai hackathon ka timing kya hai?", true},
		{"Registration link bhej do yar", true},
		// Unanswerable
		{"Is lunch free?", false},
		{"Is there parking available?", false},
		{"Will I get a certificate?", false},
		{"Can I sleep there at night?", false},
	}

	fmt.Printf("%-35s | %-15s | %-10s | %s\n", "Question", "Answerable?", "Passed?", "Output/Error")
	fmt.Println(strings.Repeat("-", 100))

	for _, q := range questions {
		prompt := fmt.Sprintf(`You are the assistant for the event described below.
Answer ONLY using the event text. 
If the answer is not clearly stated or you don't know, reply EXACTLY with this fallback message: "%s"
Keep answers under 4 sentences.
Reply in the same language/style the user wrote in (English or Hinglish).
NEVER invent times, places, prices, or rules.

Event details:
%s

Question: %s`, fallbackMsg, info, q.Q)

		ans, err := client.ChatCompletion(context.Background(), prompt)
		if err != nil {
			fmt.Printf("%-35s | %-15v | %-10s | %v\n", q.Q, q.Answerable, "FAIL", err)
			continue
		}

		passed := true
		if !q.Answerable {
			// If it's not answerable, it MUST be exactly the fallback msg
			if ans != fallbackMsg {
				passed = false
			}
		} else {
			// If it is answerable, it MUST NOT be the fallback msg
			if ans == fallbackMsg {
				passed = false
			}
		}

		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		
		fmt.Printf("%-35s | %-15v | %-10s | %s\n", q.Q, q.Answerable, status, ans)
		
		// Prevent rate limits (5 RPM on free tier)
		time.Sleep(12 * time.Second)
	}
}
