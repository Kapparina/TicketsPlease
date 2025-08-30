package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"

	"github.com/kapparina/ticketsplease/cmd"
	"github.com/kapparina/ticketsplease/cmd/commands"
	"github.com/kapparina/ticketsplease/cmd/components"
	"github.com/kapparina/ticketsplease/cmd/handlers"
)

var (
	Version = "dev"
	Commit  = "unknown"
	GitTag  = "unknown"
)

func main() {
	shouldSyncCommands := flag.Bool("sync-commands", true, "Whether to sync commands to discord")
	path := flag.String("config", "config.toml", "path to config")
	flag.Parse()
	cfg, err := cmd.LoadConfig(*path)
	if err != nil {
		slog.Error("Failed to read config", slog.Any("err", err))
		os.Exit(-1)
	}
	setupLogger(cfg.Log)
	slog.Info(
		"Starting ticket manager...",
		slog.String("version", Version),
		slog.String("git_tag", GitTag),
		slog.String("commit", Commit),
	)
	slog.Info("Command sync status", slog.Bool("sync", *shouldSyncCommands))
	b := cmd.New(*cfg, Version, Commit, GitTag)
	m := handler.New()
	m.Use(middleware.Logger)
	m.Command("/test", handlers.TestHandler)
	m.Autocomplete("/test", handlers.TestAutocompleteHandler)
	m.Command("/version", handlers.VersionHandler(b))
	m.Component("/test-button", components.TestComponent)
	m.Command("/ticket", handlers.CreateTicketHandler(b))
	// m.Autocomplete("/ticket", handlers.TicketAutocompleteHandler)
	m.Command("/help", handlers.HelpHandler(b))
	if err = b.SetupBot(m, bot.NewListenerFunc(b.OnReady), bot.NewListenerFunc(b.OnJoin), handlers.MessageHandler(b)); err != nil {
		slog.Error("Failed to setup bot", slog.Any("err", err))
		os.Exit(-1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()
	if *shouldSyncCommands {
		slog.Info(
			"Attempting to sync commands...",
			slog.Any("guild_ids", cfg.Bot.DevGuilds), slog.Any("commands", commands.Commands),
		)
		if err = handler.SyncCommands(b.Client, commands.Commands, cfg.Bot.DevGuilds); err != nil {
			slog.Warn("Failed to sync commands", slog.Any("err", err))
			slog.Info("Attempting to force register commands...")
			if _, err = b.Client.Rest().SetGlobalCommands(b.Client.ApplicationID(), commands.Commands); err != nil {
				slog.Warn("Failed to force register commands", slog.Any("err", err))
				slog.Error("Failed to update commands, continuing without syncing")
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = b.Client.OpenGateway(ctx); err != nil {
		slog.Error("Failed to open gateway", slog.Any("err", err))
		os.Exit(-1)
	}
	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
}

func setupLogger(cfg cmd.LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	var sHandler slog.Handler
	switch cfg.Format {
	case "json":
		sHandler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		sHandler = slog.NewTextHandler(os.Stdout, opts)
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(sHandler))
}
