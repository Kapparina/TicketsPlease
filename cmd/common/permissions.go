package common

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
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

func FilterRolesByPermission(roles []discord.Role, targetSubset ...PermissionSubset) []discord.Role {
	var filteredRoles []discord.Role
	for _, r := range roles {
		for _, subset := range targetSubset {
			if r.Permissions.Has(PermissionAssignments[subset]...) {
				slog.Debug(
					"Filtered role",
					slog.Any("role", r.Name),
					slog.Any("method", "permission"),
				)
				filteredRoles = append(filteredRoles, r)
			}
		}
	}
	return filteredRoles
}

func FilterRolesByNames(roles []discord.Role, targetNames ...string) []discord.Role {
	var filteredRoles []discord.Role
	for _, r := range roles {
		for _, name := range targetNames {
			if r.Name == name {
				slog.Debug(
					"Filtered role",
					slog.Any("role", r.Name),
					slog.Any("method", "name"),
				)
				filteredRoles = append(filteredRoles, r)
			}
		}
	}
	return filteredRoles
}

func FilterRolesRemoveManaged(roles []discord.Role) []discord.Role {
	var filteredRoles []discord.Role
	for _, r := range roles {
		if !r.Managed {
			slog.Debug(
				"Filtered role",
				slog.Any("role", r.Name),
				slog.Any("method", "managed"),
			)
			filteredRoles = append(filteredRoles, r)
		}
	}
	return filteredRoles
}
