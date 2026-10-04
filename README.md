# Event FAQ Bot

A Telegram bot in Go for answering questions about a real event using a predefined `event.md` and Google AI Studio's models (like Gemma or Gemini).

## Setup (Under 10 Steps)

1. Clone or download this repository.
2. Ensure you have Go 1.22+ installed.
3. Get a Telegram Bot token by talking to [@BotFather](https://t.me/botfather) on Telegram.
4. Get a Google AI Studio API key from [aistudio.google.com](https://aistudio.google.com/).
5. Copy `.env.example` to a new file named `.env`.
6. Fill out the variables in `.env` (`TELEGRAM_BOT_TOKEN`, `LLM_API_KEY`, etc.).
7. Put your event details into `event.md`.
8. Run the bot: `go run main.go`.
9. Send `/start` to your bot on Telegram and begin asking questions!

## Switching to Local Ollama

To use this bot with a local open-weight model like Gemma served by Ollama, update your `.env`:

```env
LLM_BASE_URL=http://localhost:11434/v1
LLM_API_KEY=dummy
LLM_MODEL=gemma:2b
```

Since the Completer interface uses standard OpenAI-compatible endpoints, it works perfectly with Ollama's OpenAI compatibility layer without needing to change any code.
