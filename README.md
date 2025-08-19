# DisGoTicketManager (TicketsPlease)

[![Build](https://github.com/kapparina/ticketsplease/actions/workflows/docker.yml/badge.svg)](https://github.com/kapparina/ticketsplease/actions/workflows/docker.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kapparina/ticketsplease)](https://golang.org/doc/devel/release.html)

A Discord ticket manager bot built with disgo. It helps your server members open support tickets and suggestions using slash commands. The bot automatically creates or maintains a "support-tickets" text channel and starts a private thread per ticket with sensible permission overwrites for moderation roles and restricted default (everyone) access.

Key technologies: Go, disgo (handler), paginator, TOML configuration, Docker.

## ✨ Features
- Slash commands
  - /ticket: create a private ticket thread under the support-tickets channel
  - /version: show running version and commit
  - /test: demo command with autocomplete and a demo button component
- Ticket flow
  - Ensures a text channel named "support-tickets" exists (creates or updates it)
  - Creates a private thread per ticket: `<username> - <subject> | (<category>)`
  - Adds the requesting user to the thread
- Categories (support & suggestions)
  - Predefined baseChoices, e.g. general-support, mod-support, staff-support, etc.
- Role-based permissions
  - Moderation roles (have ViewAuditLog + ManageMessages) get thread permissions
  - Everyone (guild) role is restricted from sending messages in the parent channel; threads are used for conversations
- Command sync
  - Supports rapid iteration via guild-scoped sync (dev_guilds) or global registration
- Logging with selectable level and format

## 📋 Requirements
- Go 1.24+
- Discord bot with intents enabled (Guilds, GuildMessages, MessageContent) in the Discord Developer Portal

## ⚙️ Configuration
This project uses a TOML configuration file and an environment variable for the bot token.

- Config file path flag: -config (default: config.toml)
- Sync commands flag: -sync-commands (default: true)

TOML structure:

```toml
[log]
# valid levels are "debug", "info", "warn", "error"
level = "info"
# valid formats are "text" and "json"
format = "text"
# whether to add the log source to the log message
add_source = true

[bot]
# add guild ids the commands should sync to; leave empty to sync globally
# example: dev_guilds = ["123456789012345678", "987654321098765432"]
dev_guilds = []
# optional token field present in config but NOT used at runtime by the bot
# the bot reads the token from the environment variable TicketsPleaseBotToken instead
# token = "..."
```

Environment variables:
- TicketsPleaseBotToken: the Discord bot token used by the application at runtime

Security note: Avoid committing real tokens to source control. This project reads the token from an environment variable and ignores the [bot].token field at runtime.

## 🧩 Commands
- `/ticket`
  - Options:
    - `category` (string choice; autocomplete): one of predefined categories
    - `subject` (string): 10 to 100 characters
    - `content` (string): 10 to 1000 characters
    - `attachment` (optional)
- `/version`: shows bot version and commit
- `/test`: demo command with autocomplete and a button labeled "test" (updates message on click)

## ▶️ Running locally
1. Set your bot token in the environment (PowerShell):
```powershell
$Env:TicketsPleaseBotToken = "your-bot-token"
```
2. Ensure config.toml exists (or provide a custom path). Adjust [log] and [bot.dev_guilds] as needed.
3. Run the bot:
```powershell
go run . -config config.toml -sync-commands=true
```

The bot logs startup info and attempts to sync commands to the configured dev_guilds (if any) or globally. Press CTRL-C to stop.

## 🐳 Docker
Build the image locally:
- PowerShell:
```powershell
docker build -t ticketsplease:local .
```

Run the container:
- PowerShell:
```powershell
docker run --rm -e TicketsPleaseBotToken=$Env:TicketsPleaseBotToken -v ${PWD}\config.toml:/var/lib/config.toml ticketsplease:local -config /var/lib/config.toml -sync-commands=true
```

### docker-compose
A docker-compose.yml is provided. It expects the TicketsPleaseBotToken env var in your shell environment.
- PowerShell:
```powershell
$Env:TicketsPleaseBotToken = "your-bot-token"
docker compose up -d
```

It mounts ./config.toml into the container at /var/lib/config.toml and passes flags: -config=/var/lib/config.toml --sync-commands=true

Note: The provided compose file references the image ghcr.io/kapparina/ticketsplease:main. You can modify it to use your locally built image if preferred.

## 🛠️ Development notes
- Module path and image references use github.com/kapparina/ticketsplease; keep this in mind when forking/renaming.
- Gateway intents used: Guilds, GuildMessages, MessageContent. Enable them in the Developer Portal for your bot.
- Logging is set up via slog; formats: text or json.

## 📄 License
This project is licensed under the Apache License 2.0. See the LICENSE file for details.
