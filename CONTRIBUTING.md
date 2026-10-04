# Contributing to Event FAQ Bot

First off, thank you for considering contributing to Event FAQ Bot! It's people like you that make Hacktoberfest and the open-source community such a great place.

## How Can I Contribute?

### 1. Reporting Bugs
- Use the Bug Report issue template.
- Describe the bug clearly, including steps to reproduce.

### 2. Suggesting Enhancements
- Use the Feature Request issue template.
- Explain how the enhancement would improve the bot.

### 3. Submitting Pull Requests
1. Fork the repository and create your branch from `main`.
2. Run `go test ./...` to ensure all tests pass.
3. Make sure your code follows standard Go formatting (`gofmt`).
4. Update the `README.md` if your changes add new features or environment variables.
5. Create a Pull Request using the provided PR template.

## Beginner Friendly Issues
If you're looking for something to work on, check the GitHub issues tab for labels like `good first issue` or `hacktoberfest`.

**Ideas for contributions:**
- Add a `Dockerfile` for easier deployment.
- Migrate `questions.log` from plain JSON lines to a SQLite database.
- Add support for multiple events instead of just one `event.md`.
- Improve the prompt engineering for edge cases.

Happy hacking!
