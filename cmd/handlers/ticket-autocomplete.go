package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/pkg/errors"

	"github.com/kapparina/ticketsplease/cmd/common"
	"github.com/kapparina/ticketsplease/cmd/utils"
)

// TicketAutocompleteHandler handles the autocomplete logic for ticket category options based on user input.
func TicketAutocompleteHandler(e *handler.AutocompleteEvent) error {
	var choices []discord.AutocompleteChoice
	focusedElement := e.Data.Focused().Name
	slog.Debug("Focused element", slog.Any("focused_element", focusedElement))
	input, _ := e.Data.Option(focusedElement)
	value, inputErr := getInputValue[string](input)
	if inputErr != nil {
		return errors.WithMessage(inputErr, "failed to get input value")
	}
	slog.Debug("Autocomplete input", slog.Any("input", input), slog.Any("input_value", value))
	switch focusedElement {
	case "category":
		baseChoices, choiceErr := common.GetCategoryChoices[discord.AutocompleteChoiceString]()
		if choiceErr != nil {
			return errors.WithMessage(choiceErr, "failed to get category choices")
		}
		slog.Debug("Autocomplete choices", slog.Any("choices", baseChoices))
		choicesInterface := make([]discord.AutocompleteChoice, len(baseChoices))
		for i, choice := range baseChoices {
			choicesInterface[i] = choice
		}
		if len(value) > 0 {
			filteredChoices := utils.GetFilteredAutocompleteOptions[discord.AutocompleteChoice](value, choicesInterface)
			if len(filteredChoices) == 0 {
				return fmt.Errorf("no results found for %s", value)
			}
			choices = filteredChoices
			break
		}
		choices = choicesInterface
		break
	default:
		return nil
	}
	return e.AutocompleteResult(choices)
}

// getInputValue deserializes the value of an AutocompleteOption into the specified generic type T and returns it with any error.
func getInputValue[T any](option discord.AutocompleteOption) (T, error) {
	var value T
	err := json.Unmarshal(option.Value, &value)
	return value, err
}
