package handlers

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func TestAutocompleteHandler(e *handler.AutocompleteEvent) error {
	return e.AutocompleteResult([]discord.AutocompleteChoice{
		discord.AutocompleteChoiceString{
			Name:  "1",
			Value: "1",
		},
		discord.AutocompleteChoiceString{
			Name:  "2",
			Value: "2",
		},
		discord.AutocompleteChoiceString{
			Name:  "3",
			Value: "3",
		},
	})
}
