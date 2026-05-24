# DisGoTicketManager (TicketsPlease)

[![Build](https://github.com/kapparina/ticketsplease/actions/workflows/docker.yml/badge.svg)](https://github.com/kapparina/ticketsplease/actions/workflows/docker.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kapparina/ticketsplease)](https://golang.org/doc/devel/release.html)

A Discord ticket manager bot built with [disgo](https://github.com/disgoorg/disgo). It helps your server members open support tickets and suggestions using
slash commands. The bot automatically creates or maintains a "support-tickets" text channel and starts a private thread
per ticket with sensible permission overrides for moderation roles and restricted default (everyone) access.

Key technologies: [Go](https://go.dev), [disgo](https://github.com/disgoorg/disgo), [paginator](https://github.com/disgoorg/paginator), [TOML](https://toml.io/en/) configuration, [Docker](https://www.docker.com/).

## ✨ Features

- Slash commands
    - `/ticket`: create a private ticket thread under the support-tickets channel
    - `/version`: show running version and commit
    - `/test`: demo command with autocomplete and a demo button component
- Ticket flow
    - Ensures a text channel named "support-tickets" exists (creates or updates it)
    - Creates a private thread per ticket: `<username> - <subject> | (<category>)`
    - Adds the requesting user to the thread
- Categories (support & suggestions)
    - Predefined `baseChoices`, e.g. `general-support`, `mod-support`, `staff-support`, etc.
- Role-based permissions
    - Moderation roles (have `ViewAuditLog` + `ManageMessages`) get thread permissions
    - Everyone (guild) role is restricted from sending messages in the parent channel; threads are used for
      conversations
- Command sync
    - Supports rapid iteration via guild-scoped sync (dev_guilds) or global registration
- Logging with configurable level and format

## 📋 Requirements

- Go 1.24+
- Discord bot with intents enabled (*Guilds*, *GuildMessages*, *MessageContent*) in the [Discord Developer Portal](https://discord.com/developers/applications)

## ⚙️ Configuration

This project uses a [TOML](https://toml.io/en/) configuration file and an environment variable for the bot token.

- Config file path flag: -config (default: `config.toml`)
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
# optional token field present in config but NOT used at runtime by the bot, which reads the token from the environment variable TicketsPleaseBotToken instead
# token = "..."
```

Environment variables:

- `TicketsPleaseBotToken`: the Discord bot token used by the application at runtime

Security note: Avoid committing real tokens to source control. This project reads the token from an environment variable
and ignores the `bot.token` field at runtime.

## 🧩 Commands

- `/help`: shows a help message and explains how to create a ticket
- `/ticket`
    - Options:
        - `category` (string choice; autocomplete): one of predefined categories
        - `subject` (string): 10 to 100 characters
        - `content` (string): 10 to 1000 characters
        - `attachment` (optional)
- `/version`: shows bot version, git tag (if available), and commit
- `/test`: demo command with autocomplete and a button labelled "test" (updates the message on click)

## ▶️ Running Locally

1. Set your bot token in the environment (PowerShell):
    ```powershell
    $Env:TicketsPleaseBotToken = "your-bot-token"
    ```
2. Ensure `config.toml` exists (or provide a custom path). Adjust `[log]` and `[bot.dev_guilds]` as needed.
3. Run the bot:

```powershell
go run . -config config.toml -sync-commands=true
```

The bot logs start-up info and attempts to sync commands to the configured dev_guilds (if any) or globally. Press CTRL-C
to stop.

## 🐳 Docker

Build the image locally:

- PowerShell:

```powershell
docker build -t ticketsplease:local .
```

Run the container:

- PowerShell:

```powershell
docker run --rm -e TicketsPleaseBotToken=$Env:TicketsPleaseBotToken -v ${PWD}\config.toml:/config/config.toml ticketsplease:local -config /config/config.toml -sync-commands=true
```

### Docker Compose

A [compose.yml](compose.yml) is provided. It expects the `TicketsPleaseBotToken` environment variable in your shell environment.

- PowerShell:

```powershell
$Env:TicketsPleaseBotToken = "your-bot-token"
docker compose up -d
```

It mounts `./config.toml` into the container at `/config/config.toml` and passes flags: `-config=/config/config.toml --sync-commands=true`

Note: The provided [compose](./compose.yml) file references the image [ghcr.io/kapparina/ticketsplease:main](https://github.com/users/kapparina/packages/container/package/ticketsplease). You can modify it to use your
locally built image if preferred.

## 🚀 CI/CD and Deployment

- Workflows
    - Docker build (Docker workflow): builds multi-arch images and pushes to [GHCR](https://ghcr.io). Version metadata is embedded via
      build arguments: VERSION, COMMIT (short SHA), and GIT_TAG. These are shown by `/version` and in startup logs.
    - Deploy (Deploy Discord Bot workflow): triggers on pushes to main and tags matching v*. It first runs checks (go
      build and go test).
      Only if they pass, it builds and pushes the image and then deploys via SSH to your host using
      [Docker Compose](https://docs.docker.com/compose/).
- Image tags
    - Tags include: semantic versioning (on tags), major.minor, branch names, and short SHA. You can pin `docker compose` to a specific
      tag (e.g. `v1.2.3` or `:<short-sha>`) instead of `:main`.
- Required GitHub Secrets for deployment:
    - `DISCORD_BOT_TOKEN`: bot token (written to `.env` as `TicketsPleaseBotToken`)
    - `SSH_PRIVATE_KEY`, `SSH_HOST`, `SSH_PORT`, `SSH_USERNAME`: for SSH access to the deployment host
- Required GitHub Variables for deployment:
    - `DEPLOYMENT_DIR`: the directory on the host where the bot will be deployed (e.g. `/home/user/ticketsplease`)
- Files deployed
    - [compose.yml](./compose.yml) is copied to the configured `DEPLOYMENT_DIR` on the host.
    - *.env* is created with `TicketsPleaseBotToken` and `IMAGE_TAG`.
    - The bot uses the bundled `/config/config.toml` by default (see [config.example.toml](./config.example.toml)).
      To use a custom config, place a `config.toml` in your `DEPLOYMENT_DIR`; it will be mounted into the container.

## 🛠️ Development Notes

- Module path and image references use [github.com/kapparina/ticketsplease](https://github.com/kapparina/ticketsplease); keep this in mind when forking/renaming.
- Gateway intents used: `Guilds`, `GuildMessages`, `MessageContent`. Enable them in the [Discord Developer Portal](https://discord.com/developers/applications) for your bot.
- Logging is set up via [slog](https://pkg.go.dev/log/slog); formats: text or JSON.

## 📄 Licence

Copyright 2025 Kapparina

This project is licensed under the [Apache Licence 2.0](https://www.apache.org/licenses/LICENSE-2.0). See the [LICENCE](LICENCE) file for details.
