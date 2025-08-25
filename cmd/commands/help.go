package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var Help = discord.SlashCommandCreate{
	Name:        "help",
	Description: "How to use this app",
}
