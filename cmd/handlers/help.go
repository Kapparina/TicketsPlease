package handlers

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"

	"github.com/kapparina/ticketsplease/cmd"
)

func HelpHandler(b *cmd.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		c, err := cmd.GetSupportChannel(b, e.GuildID())
		if err != nil || c == 0 {
			return errors.WithMessage(err, "failed to get support channel")
		}
		if err = sendHelpMessage(e, c); err != nil {
			return err
		}
		return nil
	}
}

func sendHelpMessage(e *handler.CommandEvent, c snowflake.ID) error {
	err := e.CreateMessage(
		discord.NewMessageCreateBuilder().
			SetContentf("Please review <#%s>", c).
			SetEphemeral(true).
			Build(),
	)
	if err != nil {
		return errors.WithMessage(err, "failed to send help message")
	}
	return nil
}
