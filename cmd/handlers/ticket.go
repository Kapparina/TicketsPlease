package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/kapparina/ticketsplease/cmd"
	"github.com/kapparina/ticketsplease/cmd/templates"
	"github.com/kapparina/ticketsplease/cmd/utils"
)

// CreateTicketHandler creates a command handler for the ticket creation command
func CreateTicketHandler(b *cmd.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		channelID, err := getOrCreateSupportChannel(b, e.GuildID())
		if err != nil {
			return err
		}
		if err = setupSupportChannel(b, &channelID, e); err != nil {
			return err
		}
		threadID, err := createTicketThread(b, channelID, e)
		if err != nil {
			return err
		}
		if err = sendTicketCreationConfirmation(e, threadID); err != nil {
			return err
		}
		return nil
	}
}

func getSupportChannelOverrides(b *cmd.Bot, guildID snowflake.ID) []discord.PermissionOverwrite {
	var overrides discord.PermissionOverwrites
	roles, roleErr := b.Client.Rest().GetRoles(guildID)
	if roleErr != nil {
		return nil
	}
	filteredRoles := utils.FilterRoles(roles, utils.Moderation)
	for _, r := range filteredRoles {
		o := discord.RolePermissionOverwrite{
			RoleID: r,
			Allow:  discord.PermissionsAllThread,
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
func getOrCreateSupportChannel(b *cmd.Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	channels, err := b.Client.Rest().GetGuildChannels(*guildID)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to get guild channels")
	}
	for _, c := range channels {
		if c.Name() == cmd.SupportChannelName {
			return updateSupportChannel(b, c.ID(), guildID)
		}
	}
	return createSupportChannel(b, guildID)
}

// createSupportChannel creates a new support channel
func createSupportChannel(b *cmd.Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	supportChannelOverrides := getSupportChannelOverrides(b, *guildID)
	c, err := b.Client.Rest().CreateGuildChannel(
		*guildID,
		discord.GuildTextChannelCreate{
			Name:                 cmd.SupportChannelName,
			Topic:                cmd.SupportChannelTopic,
			PermissionOverwrites: supportChannelOverrides,
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to create support-tickets channel")
	}
	return c.ID(), nil
}

// updateSupportChannel updates an existing support channel
func updateSupportChannel(b *cmd.Bot, channelID snowflake.ID, guildID *snowflake.ID) (snowflake.ID, error) {
	supportChannelOverrides := getSupportChannelOverrides(b, *guildID)
	targetChannelType := discord.ChannelTypeGuildText
	c, err := b.Client.Rest().UpdateChannel(
		channelID,
		discord.GuildTextChannelUpdate{
			Name:                 &cmd.SupportChannelName,
			Type:                 &targetChannelType,
			Topic:                &cmd.SupportChannelTopic,
			PermissionOverwrites: &supportChannelOverrides,
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to update support-tickets channel")
	}
	return c.ID(), nil
}

func setupSupportChannel(b *cmd.Bot, c *snowflake.ID, e *handler.CommandEvent) error {
	if err := deleteExistingMessages(b, c); err != nil {
		return err
	}
	commandName := e.SlashCommandInteractionData().CommandName()
	helpData := templates.HelpData{CommandName: commandName}
	content, err := templates.PopulateHelpData(helpData)
	if err != nil {
		return err
	}
	if err = postHelpMessage(b, c, content); err != nil {
		return err
	}
	return nil
}

func postHelpMessage(b *cmd.Bot, c *snowflake.ID, content string) error {
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

func deleteExistingMessages(b *cmd.Bot, c *snowflake.ID) error {
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

func getExistingBotMessages(b *cmd.Bot, c *snowflake.ID) ([]discord.Message, error) {
	messages, err := b.Client.Rest().GetMessages(*c, 0, 0, 0, 100)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get existing support channel messages")
	}
	return messages, nil
}

// createTicketThread creates a private thread for the ticket
func createTicketThread(b *cmd.Bot, channelID snowflake.ID, e *handler.CommandEvent) (snowflake.ID, error) {
	data := e.SlashCommandInteractionData()
	t, err := b.Client.Rest().CreateThread(
		channelID,
		discord.GuildPrivateThreadCreate{
			Name: fmt.Sprintf(
				"%s - %s | (%s)",
				e.User().Username, data.String("subject"), data.String("category"),
			),
			AutoArchiveDuration: 60,
		},
	)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to create thread")
	}
	if err = b.Client.Rest().AddThreadMember(
		t.ID(),
		e.User().ID,
	); err != nil {
		return 0, errors.WithMessage(err, "failed to add thread member")
	}
	if err = sendTicketContent(b, t.ID(), e); err != nil {
		return 0, errors.WithMessage(err, "failed to send ticket content")
	}
	return t.ID(), nil
}

// sendTicketCreationConfirmation sends a confirmation message to the user
func sendTicketCreationConfirmation(e *handler.CommandEvent, threadID snowflake.ID) error {
	err := e.CreateMessage(
		discord.NewMessageCreateBuilder().
			SetContentf("Created ticket: <#%s>", threadID).
			SetEphemeral(true).
			Build(),
	)
	if err != nil {
		return errors.WithMessage(err, "failed to send confirmation message")
	}
	return nil
}

//goland:noinspection StructuralWrap
func determineRoleFilter(category cmd.Category) []utils.PermissionSubset {
	var subsets []utils.PermissionSubset
	if category.RequiresMod() {
		subsets = append(subsets, utils.Moderation)
	}
	if category.RequiresAdmin() || category.RequiresStaff() || category.RequiresOwner() { // TODO: add staff & owner subsets
		subsets = append(subsets, utils.Administration)
	}
	if len(subsets) == 0 {
		subsets = append(subsets, utils.Moderation)
	}
	return subsets
}

func populateTicketContent(b *cmd.Bot, e *handler.CommandEvent) (string, error) {
	data := e.SlashCommandInteractionData()
	roles, err := b.Client.Rest().GetRoles(*e.GuildID())
	if err != nil {
		slog.Error("Failed to get roles", slog.Any("err", err))
	}
	category, _ := cmd.FindCategoryByDescription(data.String("category"))
	filterSubsets := determineRoleFilter(category)
	filteredRoles := utils.FilterRoles(roles, filterSubsets...)
	moderatorRoleIDs := make([]string, len(filteredRoles))
	for i, r := range filteredRoles {
		moderatorRoleIDs[i] = r.String()
	}

	// Get attachment URL if present
	var attachmentURL string
	if att, ok := data.OptAttachment("attachment"); ok {
		attachmentURL = att.URL
	}

	ticketData := templates.TicketData{
		Category:      data.String("category"),
		Username:      e.User().Username,
		Subject:       data.String("subject"),
		Content:       data.String("content"),
		Moderators:    moderatorRoleIDs,
		AttachmentURL: attachmentURL,
	}
	return templates.PopulateTicketData(ticketData)
}

func sendTicketContent(b *cmd.Bot, threadID snowflake.ID, e *handler.CommandEvent) error {
	content, err := populateTicketContent(b, e)
	if err != nil {
		return errors.WithMessage(err, "failed to populate ticket content")
	}

	_, err = b.Client.Rest().CreateMessage(
		threadID,
		discord.NewMessageCreateBuilder().
			SetContent(content).
			Build(),
	)
	if err != nil {
		return errors.WithMessage(err, "failed to create message in thread")
	}

	return nil
}
