package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var test = discord.SlashCommandCreate{
	Name:        "test",
	Description: "test command",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:         "choice",
			Description:  "some autocomplete choice",
			Required:     true,
			Autocomplete: true,
		},
	},
}
