package utils

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

type PermissionSubset int

const (
	Moderation PermissionSubset = iota
	Vanity
	Administration
)

var PermissionAssignments = map[PermissionSubset][]discord.Permissions{
	Moderation: {
		discord.PermissionViewAuditLog,
		discord.PermissionManageMessages,
	},
	Vanity: nil,
	Administration: {
		discord.PermissionAdministrator,
	},
}

func FilterRoles(roles []discord.Role, targetSubset ...PermissionSubset) []snowflake.ID {
	var filteredRoles []snowflake.ID
	for _, r := range roles {
		for _, subset := range targetSubset {
			if r.Permissions.Has(PermissionAssignments[subset]...) {
				filteredRoles = append(filteredRoles, r.ID)
			}
		}
	}
	return filteredRoles
}
