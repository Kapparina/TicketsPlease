package common

import (
	"strings"
)

type ChoiceOption interface {
	ChoiceName() string
}

func GetFilteredAutocompleteOptions[T ChoiceOption](input string, options []T) []T {
	var filteredChoices []T
	for i, c := range options {
		if strings.Contains(c.ChoiceName(), input) {
			filteredChoices = append(filteredChoices, options[i])
		}
	}
	return filteredChoices
}
