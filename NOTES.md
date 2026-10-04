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

## Round 2 fixes

### Bugs Fixed
1. **Missing /start reply (Rate limiter bug)**: 
   - *How it was reproduced*: Sending `/help` or `/start` immediately after a normal question dropped the message silently. 
   - *What changed*: Commands now bypass the rate limiter completely. Also added a warning message for rate-limited text messages (throttled to 1 warning per 10s per user).
2. **Contact Validation**: 
   - *What changed*: Added a startup check in `main.go` that strictly validates `ORGANIZER_CONTACT`. It will `log.Fatal` if empty or if it contains `example.com`.
3. **Non-text input**: 
   - *What changed*: Handled non-text updates (photos, stickers) safely, returning a friendly "I can only read text messages" reply without crashing.

### Prompt Hardening & New Markers
Added strict rules to output specific text markers instead of LLM-generated responses for certain intents:
- `ALLDETAILS|EN` or `ALLDETAILS|HI` for full event summaries (handled cleanly by code without Markdown).
- `NOINFO|EN` or `NOINFO|HI` for queries missing from `event.md`. 
- `OFFTOPIC|EN` or `OFFTOPIC|HI` for jokes, prompt extraction, or ignoring instructions.

### Live Testing & Rate Limits
Attempted to run the 19 test cases against `gemini-2.5-flash` again using the new prompt rules.
However, the Google AI Studio free tier enforces a strict **20 requests per day** limit (`GenerateRequestsPerDayPerProjectPerModel-FreeTier`), which was exhausted by the previous evaluation script + the first 2 questions of this run. 

*Why it can't be re-run right now*: The API returns a 429 quota exhausted error asking to retry in 14 hours. 

Here is what the prompt rules are designed to yield:
- **Wrong Premise (Before)**: "Yes, the hackathon is on the 26th."
- **Wrong Premise (After)**: "The Tech Innovators Hackathon 2026 is actually taking place on October 24-25, 2026."
- **Off-topic (Before)**: "Here is a Python function: ..." 
- **Off-topic (After)**: Returns `OFFTOPIC|EN`, triggering code fallback: "I can only help with questions about this event. Try /help to see what I can answer."
- **Prompt extraction (Before)**: "My instructions are..." 
- **Prompt extraction (After)**: Returns `OFFTOPIC|EN`.
