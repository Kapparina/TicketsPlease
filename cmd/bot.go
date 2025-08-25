package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
)

func New(cfg Config, version, commit, tag string) *Bot {
	return &Bot{
		Cfg:       cfg,
		Paginator: paginator.New(),
		Version:   version,
		Commit:    commit,
		GitTag:    tag,
	}
}

type Bot struct {
	Cfg       Config
	Client    bot.Client
	Paginator *paginator.Manager
	Version   string
	Commit    string
	GitTag    string
}

func (b *Bot) SetupBot(listeners ...bot.EventListener) error {
	client, err := disgo.New(
		os.Getenv("TicketsPleaseBotToken"),
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentsGuild, gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListeners(b.Paginator),
		bot.WithEventListeners(listeners...),
	)
	if err != nil {
		return err
	}
	b.Client = client
	return nil
}

func (b *Bot) OnReady(e *events.Ready) {
	slog.Info("Setting presence...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Client.SetPresence(ctx, gateway.WithListeningActivity("you"), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		slog.Error("Failed to set presence", slog.Any("err", err))
	}
	slog.Info("Setting up support channel...")
	var guildIDs []snowflake.ID
	for _, g := range e.Guilds {
		guildIDs = append(guildIDs, g.ID)
	}
	if err := ConfigureSupportChannel(ctx, b, guildIDs...); err != nil {
		slog.Error("Failed to configure support channel", slog.Any("err", err))
	}
	slog.Info("Bot ready!")
}

func (b *Bot) OnJoin(e *events.GuildJoin) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	slog.Info("Setting up support channel...")
	if err := ConfigureSupportChannel(ctx, b, e.GuildID); err != nil {
		slog.Error("Failed to configure support channel", slog.Any("err", err))
	}
}
