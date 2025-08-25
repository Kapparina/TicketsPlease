package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/kapparina/ticketsplease/cmd/commands"
	"github.com/kapparina/ticketsplease/cmd/templates"
	"github.com/kapparina/ticketsplease/cmd/utils"
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
		eg, ctx := errgroup.WithContext(ctx)
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
				return err
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
	filteredRoles := utils.FilterRolesByPermission(roles, utils.Moderation)
	slog.Debug("Filtered roles", slog.Any("filtered_roles", filteredRoles))
	for _, r := range filteredRoles {
		o := discord.RolePermissionOverwrite{
			RoleID: r,
			Allow:  discord.PermissionsAllThread,
			Deny:   discord.PermissionsNone,
		}
		overrides = append(overrides, o)
	}
	filteredRoles = utils.FilterRolesByNames(roles, b.Cfg.Bot.Name)
	slog.Debug("Filtered roles", slog.Any("filtered_roles", filteredRoles))
	for _, r := range filteredRoles {
		o := discord.RolePermissionOverwrite{
			RoleID: r,
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
	if err := deleteExistingMessages(b, c); err != nil {
		return err
	}
	helpData := templates.HelpData{CommandName: cmdName, Version: b.GitTag}
	content, err := templates.PopulateHelpData(helpData)
	if err != nil {
		return err
	}
	if err = PostHelpMessage(b, c, content); err != nil {
		return err
	}
	return nil
}

// PostHelpMessage sends a message with the given content to a specified Discord channel
// using the provided bot instance.
func PostHelpMessage(b *Bot, c *snowflake.ID, content string) error {
	_, err := b.Client.Rest().CreateMessage(
		*c,
		discord.NewMessageCreateBuilder().
			SetContent(content).
			Build(),
	)
	if err != nil {
		return errors.WithMessage(err, "failed to create help message")
	}
	return nil
}

// deleteExistingMessages removes all existing bot messages from the specified Discord channel.
// It retrieves messages using the bot client and deletes them in parallel with a concurrency limit.
func deleteExistingMessages(b *Bot, c *snowflake.ID) error {
	messages, err := getExistingBotMessages(b, c)
	if err != nil {
		return err
	}
	deleteMessages := func(ctx context.Context, messages []discord.Message) error {
		eg, ctx := errgroup.WithContext(ctx)
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
	if err = deleteMessages(parentCtx, messages); err != nil {
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
