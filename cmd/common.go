package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/kapparina/ticketsplease/cmd/commands"
	"github.com/kapparina/ticketsplease/cmd/common"
	"github.com/kapparina/ticketsplease/cmd/templates"
)

// Support channel constants
var (
	SupportChannelName  = "support-tickets"
	SupportChannelTopic = "Support tickets & suggestions"
)

// GetSupportChannel retrieves the ID of the support channel with the specified name in the given guild.
// Returns the channel ID if found, otherwise returns 0 and an error if an issue occurs or the channel does not exist.
func GetSupportChannel(b *Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	channels, err := b.Client.Rest().GetGuildChannels(*guildID)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to get guild channels")
	}
	for _, c := range channels {
		if c.Name() == SupportChannelName {
			return c.ID(), nil
		}
	}
	return 0, nil
}

func ConfigureSupportChannel(ctx context.Context, b *Bot, guilds ...snowflake.ID) error {
	configFunc := func(ctx context.Context, guilds []snowflake.ID) error {
		eg, _ := errgroup.WithContext(ctx)
		eg.SetLimit(10)
		for _, g := range guilds {
			slog.Info("Setting up support channel for guild", slog.Any("guild_id", g))
			currentGuild := g
			if b.Client.Caches().IsGuildUnavailable(currentGuild) {
				slog.Warn("Guild is unavailable", slog.Any("guild_id", currentGuild))
				continue
			}
			eg.TryGo(func() error {
				c, err := getOrCreateSupportChannel(b, &currentGuild)
				if err != nil {
					return err
				}
				if err = setupSupportChannel(b, &c, commands.Ticket.CommandName()); err != nil {
					return err
				}
				slog.Info("Support channel setup successful", slog.Any("guild_id", currentGuild))
				return nil
			})
		}
		return eg.Wait()
	}
	parentCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := configFunc(parentCtx, guilds); err != nil {
		return errors.WithMessage(err, "failed to configure support channel")
	}
	return nil
}

// getSupportChannelOverrides generates a list of permission overrides for a support channel in a specific guild.
// It filters roles based on the Moderation subset and assigns specified permissions to the generated overrides.
func getSupportChannelOverrides(b *Bot, guildID snowflake.ID) []discord.PermissionOverwrite {
	var overrides discord.PermissionOverwrites
	roles, roleErr := b.Client.Rest().GetRoles(guildID)
	if roleErr != nil {
		return nil
	}
	filteredRoles := common.FilterRolesByPermission(roles, common.Moderation)
	slog.Debug("Filtered roles", slog.Any("filtered_roles", filteredRoles))
	for _, r := range filteredRoles {
		o := discord.RolePermissionOverwrite{
			RoleID: r.ID,
			Allow:  discord.PermissionsAllThread,
			Deny:   discord.PermissionsNone,
		}
		overrides = append(overrides, o)
	}
	filteredRoles = common.FilterRolesByNames(roles, b.Cfg.Bot.Name)
	slog.Debug("Filtered roles", slog.Any("filtered_roles", filteredRoles))
	for _, r := range filteredRoles {
		o := discord.RolePermissionOverwrite{
			RoleID: r.ID,
			Allow:  discord.PermissionsAllChannel | discord.PermissionsAllThread,
			Deny:   discord.PermissionsNone,
		}
		overrides = append(overrides, o)
	}
	overrides = append(overrides, discord.RolePermissionOverwrite{
		RoleID: guildID,
		Allow: discord.PermissionSendMessagesInThreads |
			discord.PermissionViewChannel |
			discord.PermissionReadMessageHistory,
		Deny: discord.PermissionReadMessageHistory |
			discord.PermissionManageThreads |
			discord.PermissionCreatePublicThreads |
			discord.PermissionCreatePrivateThreads |
			discord.PermissionSendMessages,
	})
	return overrides
}

// getOrCreateSupportChannel finds or creates the support channel
func getOrCreateSupportChannel(b *Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	c, err := GetSupportChannel(b, guildID)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to get support channel")
	}
	if c != 0 {
		return updateSupportChannel(b, c, guildID)
	} else {
		return createSupportChannel(b, guildID)
	}
}

// createSupportChannel creates a new support channel
func createSupportChannel(b *Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	supportChannelOverrides := getSupportChannelOverrides(b, *guildID)
	c, err := b.Client.Rest().CreateGuildChannel(
		*guildID,
		discord.GuildTextChannelCreate{
			Name:                 SupportChannelName,
			Topic:                SupportChannelTopic,
			PermissionOverwrites: supportChannelOverrides,
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to create support-tickets channel")
	}
	return c.ID(), nil
}

// updateSupportChannel updates an existing support channel
func updateSupportChannel(b *Bot, channelID snowflake.ID, guildID *snowflake.ID) (snowflake.ID, error) {
	supportChannelOverrides := getSupportChannelOverrides(b, *guildID)
	targetChannelType := discord.ChannelTypeGuildText
	c, err := b.Client.Rest().UpdateChannel(
		channelID,
		discord.GuildTextChannelUpdate{
			Name:                 &SupportChannelName,
			Type:                 &targetChannelType,
			Topic:                &SupportChannelTopic,
			PermissionOverwrites: &supportChannelOverrides,
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to update support-tickets channel")
	}
	return c.ID(), nil
}

// setupSupportChannel prepares a Discord support channel by clearing existing messages and posting a help message.
func setupSupportChannel(b *Bot, c *snowflake.ID, cmdName string) error {
	baseContent, err := templates.PopulateHelpData(templates.HelpData{CommandName: cmdName, Version: b.GitTag})
	if err != nil {
		return errors.WithMessage(err, "failed to populate base help message")
	}
	messages, err := getExistingBotMessages(b, c)
	if err != nil {
		return err
	}
	l := len(messages)
	lastMessage := messages[l-1]
	botUser, err := b.Client.Rest().GetCurrentUser("")
	if err != nil {
		return errors.WithMessage(err, "failed to get current user")
	}
	botUserId := botUser.ID
	if messages[l-1].Content == baseContent {
		slog.Info("Support channel is already configured")
		return nil
	} else if l >= 1 && lastMessage.Author.ID == botUserId {
		slog.Info("Last message is from the bot, attempting to update...")
		if _, err = b.Client.Rest().UpdateMessage(*c, lastMessage.ID, discord.NewMessageUpdateBuilder().
			SetContent(baseContent).
			Build(),
		); err != nil {
			slog.Error("Failed to update last message", slog.Any("err", err))
			slog.Info("Attempting to overwrite existing messages...")
			if err = deleteExistingMessages(b, c, messages); err != nil {
				return err
			}
		} else {
			return nil
		}
		slog.Info("Deletion successful, attempting to send new message...")
		_, err = b.Client.Rest().CreateMessage(
			*c,
			discord.NewMessageCreateBuilder().
				SetContent(baseContent).
				Build(),
		)
		if err != nil {
			return errors.WithMessage(err, "failed to send help message")
		}
	}
	return nil
}

// PostHelpMessage sends a message with the given content to a specified Discord channel
// using the provided bot instance.
func PostHelpMessage(b *Bot, c *snowflake.ID, data templates.HelpData, e *handler.CommandEvent) error {
	var content string
	var err error
	if e != nil {
		content, err = templates.PopulateEphemeralHelpData(data)
	} else {
		content, err = templates.PopulateHelpData(data)
	}
	if err != nil {
		return err
	}
	if e != nil {
		err = e.CreateMessage(
			discord.NewMessageCreateBuilder().
				SetContent(content).
				SetEphemeral(true).
				Build(),
		)
	} else {
		_, err = b.Client.Rest().CreateMessage(
			*c,
			discord.NewMessageCreateBuilder().
				SetContent(content).
				Build(),
		)
	}
	if err != nil {
		return errors.WithMessage(err, "failed to create help message")
	}
	return nil
}

// deleteExistingMessages removes all existing bot messages from the specified Discord channel.
// It retrieves messages using the bot client and deletes them in parallel with a concurrency limit.
func deleteExistingMessages(b *Bot, c *snowflake.ID, messages []discord.Message) error {
	deleteMessages := func(ctx context.Context, messages []discord.Message) error {
		eg, _ := errgroup.WithContext(ctx)
		eg.SetLimit(10)
		for _, m := range messages {
			currentMessage := m
			eg.TryGo(func() error {
				return b.Client.Rest().DeleteMessage(*c, currentMessage.ID)
			})
		}
		return eg.Wait()
	}
	parentCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := deleteMessages(parentCtx, messages); err != nil {
		return errors.WithMessage(err, "failed to delete existing messages")
	}
	return nil
}

// getExistingBotMessages retrieves up to 100 messages from the specified Discord channel using the provided bot client.
// Returns a slice of messages or an error if the retrieval fails.
func getExistingBotMessages(b *Bot, c *snowflake.ID) ([]discord.Message, error) {
	messages, err := b.Client.Rest().GetMessages(*c, 0, 0, 0, 100)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get existing support channel messages")
	}
	return messages, nil
}
