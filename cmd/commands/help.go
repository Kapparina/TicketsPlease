package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var help = discord.SlashCommandCreate{
	Name:        "help",
	Description: "How to use this app",
}
