package commands

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"

	"github.com/kapparina/ticketsplease/cmd/common"
)

var (
	MinTicketSubjectLength    = 3
	MinTicketSubjectLengthPtr = &MinTicketSubjectLength
	MaxTicketSubjectLength    = 100
	MaxTicketSubjectLengthPtr = &MaxTicketSubjectLength
	MaxTicketContentLength    = 1000
	MaxTicketContentLengthPtr = &MaxTicketContentLength
)

var Ticket = discord.SlashCommandCreate{
	Name:        "ticket",
	Description: "Create a ticket",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:         "category",
			Description:  "The category of the ticket",
			Required:     true,
			Autocomplete: false,
			Choices: func() []discord.ApplicationCommandOptionChoiceString {
				choices, err := common.GetCategoryChoices[discord.ApplicationCommandOptionChoiceString]()
				if err != nil {
					slog.Error("Failed to get category choices", slog.Any("err", err))
				}
				return choices
			}(),
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
