package commands

import (
	"github.com/disgoorg/disgo/discord"
)

var version = discord.SlashCommandCreate{
	Name:        "version",
	Description: "version command",
}
