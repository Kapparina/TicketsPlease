package cmd

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/pkg/errors"
)

// Support channel constants
var (
	SupportChannelName  = "support-tickets"
	SupportChannelTopic = "Support tickets & suggestions"
)

// GetSupportChannel retrieves the ID of the support channel with the specified name in the given guild.
// Returns the channel ID if found, otherwise returns 0 and an error if an issue occurs or the channel does not exist.
func GetSupportChannel(b *Bot, guildID *snowflake.ID) (snowflake.ID, error) {
	channels, err := b.Client.Rest().GetGuildChannels(*guildID)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to get guild channels")
	}
	for _, c := range channels {
		if c.Name() == SupportChannelName {
			return c.ID(), nil
		}
	}
	return 0, nil
}
