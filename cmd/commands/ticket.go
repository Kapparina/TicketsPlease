package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var (
	MinTicketSubjectLength    = 5
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
			Autocomplete: true,
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
