package handlers

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func TestHandler(e *handler.CommandEvent) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("test command. Choice: %s", e.SlashCommandInteractionData().String("choice")).
		AddActionRow(discord.NewPrimaryButton("test", "/test-button")).
		SetEphemeral(true).
		Build(),
	)
}
