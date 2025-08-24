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
