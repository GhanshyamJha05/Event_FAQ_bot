# Project Notes & Testing

## What Was Tested
- **Prompt Construction & Rate Limiter:** Tested locally using Go unit tests and a fake LLM completer.
- **Thought Stripping:** Tested edge cases for handling `<thought>...</thought>` tags from models like Gemma 4, ensuring reasoning isn't leaked to users or logs, and handling unclosed tags as errors.
- **LLM Real-world Evaluation:** Sent 10 real questions (6 answerable, 4 unanswerable) to Gemini 2.5 Flash / Gemma to test prompt robustness. 

## Model Quirks & Fixes
- **Initial Rate Limits:** Testing 10 questions at once rapidly triggered Google AI Studio's free tier rate limits (429 Too Many Requests - 5 requests per minute limit). This was fixed by adding a 12-second backoff between requests in the evaluation script.
- **System/Developer Roles:** AI Studio's OpenAI endpoint often rejects separate system/developer roles for certain models like Gemma. The prompt construction was fixed by bundling all instructions and the event details into a single `user` message.
- **Reasoning Tokens:** Models returning reasoning inside `<thought>` tags require careful text parsing (implemented in `client.go`). Setting `max_tokens` to `1024` was crucial because the thinking tokens count towards the overall limit, causing the actual response to truncate if set too low.

## Final Prompt Efficacy
The prompt successfully contained the LLM to only use `event.md`. Unanswerable questions like "Is lunch free?", "Is there parking available?", and "Will I get a certificate?" correctly triggered the fallback message without inventing any information.
