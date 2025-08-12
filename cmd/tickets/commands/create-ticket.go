package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"

	"github.com/kapparina/ticketsplease/cmd/tickets"
)

var (
	MinTicketSubjectLength    = 10
	MinTicketSubjectLengthPtr = &MinTicketSubjectLength
	MaxTicketSubjectLength    = 100
	MaxTicketSubjectLengthPtr = &MaxTicketSubjectLength
	MaxTicketContentLength    = 1000
	MaxTicketContentLengthPtr = &MaxTicketContentLength
)

var createTicket = discord.SlashCommandCreate{
	Name:        "ticket",
	Description: "Create a ticket",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "category",
			Description: "The category of the ticket",
			Required:    true,
			Choices: func() []discord.ApplicationCommandOptionChoiceString {
				var choices []discord.ApplicationCommandOptionChoiceString
				for _, category := range tickets.Categories {
					choices = append(choices, discord.ApplicationCommandOptionChoiceString{
						Name:  category.Title,
						Value: category.Description,
					})
				}
				return choices
			}(),
			Autocomplete: false,
		},
		discord.ApplicationCommandOptionString{
			Name:        "subject",
			Description: "A brief description of the ticket",
			Required:    true,
			MinLength:   MinTicketSubjectLengthPtr,
			MaxLength:   MaxTicketSubjectLengthPtr,
		},
		discord.ApplicationCommandOptionString{
			Name:        "content",
			Description: "Broader specifics of the ticket",
			Required:    true,
			MinLength:   MinTicketSubjectLengthPtr,
			MaxLength:   MaxTicketContentLengthPtr,
		},
		discord.ApplicationCommandOptionAttachment{
			Name:        "attachment",
			Description: "An optional attachment to send with the ticket",
			Required:    false,
		},
	},
}

func CreateTicketHandler(b *tickets.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		data := e.SlashCommandInteractionData()
		channels, err := b.Client.Rest().GetGuildChannels(*e.GuildID())
		if err != nil {
			return errors.WithMessage(err, "failed to get guild channels")
		}
		var channelID snowflake.ID
		for _, c := range channels {
			if c.Name() == "support-tickets" {
				channelID = c.ID()
			}
		}
		if channelID == 0 {
			c, channelErr := b.Client.Rest().CreateGuildChannel(
				*e.GuildID(),
				discord.GuildTextChannelCreate{
					Name:  "support-tickets",
					Topic: "Support tickets & suggestions",
				},
			)
			if channelErr != nil {
				return errors.WithMessage(channelErr, "failed to create support-tickets channel")
			}
			channelID = c.ID()
		}
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
			return errors.WithMessage(err, "failed to create thread")
		}
		if err = b.Client.Rest().AddThreadMember(
			t.ID(),
			e.User().ID,
		); err != nil {
			return errors.WithMessage(err, "failed to add thread member")
		}
		if err = e.CreateMessage(
			discord.NewMessageCreateBuilder().
				SetContentf("Created ticket: <#%s>", t.ID()).
				SetEphemeral(true).
				Build(),
		); err != nil {
			return errors.WithMessage(err, "failed to send message")
		}
		return nil
	}
}
