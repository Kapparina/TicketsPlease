package handlers

import (
	"github.com/disgoorg/disgo/handler"

	"github.com/kapparina/ticketsplease/cmd"
	"github.com/kapparina/ticketsplease/cmd/commands"
	"github.com/kapparina/ticketsplease/cmd/templates"
)

func HelpHandler(b *cmd.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		helpData := templates.HelpData{CommandName: commands.Ticket.CommandName()}
		if err := cmd.PostHelpMessage(b, nil, helpData, e); err != nil {
			return err
		}
		return nil
	}
}
