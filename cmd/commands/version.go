package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/kapparina/ticketsplease/cmd"
)

var version = discord.SlashCommandCreate{
	Name:        "version",
	Description: "version command",
}

func VersionHandler(b *cmd.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(
			discord.NewMessageCreateBuilder().
				SetContentf("Version: %s\nCommit: %s\nVersion Tag: %s", b.Version, b.Commit, b.GitTag).
				SetEphemeral(true).
				Build(),
		)
	}
}
