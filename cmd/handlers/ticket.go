package handlers

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"

	"github.com/kapparina/ticketsplease/cmd"
	"github.com/kapparina/ticketsplease/cmd/common"
	"github.com/kapparina/ticketsplease/cmd/templates"
	"github.com/kapparina/ticketsplease/cmd/utils"
)

// CreateTicketHandler creates a command handler for the ticket creation command
func CreateTicketHandler(b *cmd.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		channelID, err := cmd.GetSupportChannel(b, e.GuildID())
		if err != nil {
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
func determineRoleFilter(category common.Category) []utils.PermissionSubset {
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

// populateTicketContent generates a formatted string for a ticket using user input and guild role data.
// It fetches custom role filters based on the ticket category and applies them to assemble moderator role IDs.
// Handles optional attachment URLs and incorporates ticket data into a predefined template.
// Returns the finalised ticket content string and any error encountered during template population.
func populateTicketContent(b *cmd.Bot, e *handler.CommandEvent) (string, error) {
	data := e.SlashCommandInteractionData()
	roles, err := b.Client.Rest().GetRoles(*e.GuildID())
	if err != nil {
		slog.Error("Failed to get roles", slog.Any("err", err))
	}
	category, _ := common.FindCategoryByDescription(data.String("category"))
	filterSubsets := determineRoleFilter(category)
	filteredRoles := utils.FilterRolesByPermission(roles, filterSubsets...)
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

// sendTicketContent sends the ticket content to the specified thread ID in the provided bot context.
// It uses populateTicketContent to generate the ticket content and handles errors during content population or message creation.
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
