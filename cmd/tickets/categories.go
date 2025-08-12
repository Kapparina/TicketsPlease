package tickets

type TicketCategory struct {
	Title       string
	Description string
}

//goland:noinspection GoCommentStart
const (
	// Support
	CategoryGeneralSupport int = iota
	CategoryModSupport
	CategoryStaffSupport
	CategoryAdminSupport
	CategoryOwnerSupport
	CategoryUserSupport

	// Suggestions
	CategoryGeneralSuggestion
	CategoryModSuggestion
	CategoryStaffSuggestion
	CategoryAdminSuggestion
	CategoryOwnerSuggestion
	CategoryUserSuggestion
)

var Categories = map[int]TicketCategory{
	CategoryGeneralSupport: {
		Title:       "general-support",
		Description: "General support questions",
	},
	CategoryModSupport: {
		Title:       "mod-support",
		Description: "Moderation support questions",
	},
	CategoryStaffSupport: {
		Title:       "staff-support",
		Description: "Staff support questions",
	},
	CategoryAdminSupport: {
		Title:       "admin-support",
		Description: "Admin support questions",
	},
	CategoryOwnerSupport: {
		Title:       "owner-support",
		Description: "Owner support questions",
	},
	CategoryUserSupport: {
		Title:       "user-support",
		Description: "User support questions",
	},
	CategoryGeneralSuggestion: {
		Title:       "general-suggestion",
		Description: "General suggestion",
	},
	CategoryModSuggestion: {
		Title:       "mod-suggestion",
		Description: "Mod suggestion",
	},
	CategoryStaffSuggestion: {
		Title:       "staff-suggestion",
		Description: "Staff suggestion",
	},
	CategoryAdminSuggestion: {
		Title:       "admin-suggestion",
		Description: "Admin suggestion",
	},
	CategoryOwnerSuggestion: {
		Title:       "owner-suggestion",
		Description: "Owner suggestion",
	},
	CategoryUserSuggestion: {
		Title:       "user-suggestion",
		Description: "User suggestion",
	},
}
