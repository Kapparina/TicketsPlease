package handlers

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/pkg/errors"

	"github.com/kapparina/ticketsplease/cmd"
)

// TicketAutocompleteHandler handles the autocomplete logic for ticket category options based on user input.
func TicketAutocompleteHandler(e *handler.AutocompleteEvent) error {
	var baseChoices []discord.AutocompleteChoice
	for _, category := range cmd.Categories {
		baseChoices = append(
			baseChoices, discord.AutocompleteChoiceString{
				Name:  category.Title,
				Value: category.Description,
			})
	}
	data := e.Data
	input, ok := data.Option("category")
	if ok {
		value, err := getInputValue[string](input)
		if err != nil {
			return errors.WithMessage(err, "failed to get input value")
		}
		slog.Debug("Autocomplete input", slog.Any("input", input), slog.Any("input_value", value))
		if len(value) > 0 {
			var choices []discord.AutocompleteChoice
			for i, c := range baseChoices {
				if strings.Contains(c.ChoiceName(), value) {
					choices = append(choices, baseChoices[i])
				}
			}
			return e.AutocompleteResult(choices)
		}
	}
	return e.AutocompleteResult(baseChoices)
}

// getInputValue deserializes the value of an AutocompleteOption into the specified generic type T and returns it with any error.
func getInputValue[T any](option discord.AutocompleteOption) (T, error) {
	var value T
	err := json.Unmarshal(option.Value, &value)
	return value, err
}
