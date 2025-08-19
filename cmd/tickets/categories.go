package tickets

type TicketCategory struct {
	Title       string
	Description string
}

type Category int

func (c Category) RequiresMod() bool {
	return c < CategoryStaffSupport
}

func (c Category) RequiresStaff() bool {
	return c < CategoryAdminSupport
}

func (c Category) RequiresAdmin() bool {
	return c < CategoryOwnerSupport
}

func (c Category) RequiresOwner() bool {
	return c >= CategoryAdminSupport
}

//goland:noinspection GoCommentStart
const (
	CategoryGeneralSuggestion Category = iota
	CategoryGeneralSupport
	CategoryUserSupport
	CategoryUserSuggestion
	CategoryModSuggestion
	CategoryModSupport
	CategoryStaffSuggestion
	CategoryStaffSupport
	CategoryAdminSuggestion
	CategoryAdminSupport
	CategoryOwnerSuggestion
	CategoryOwnerSupport
)

var Categories = map[Category]TicketCategory{
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

func FindCategoryByDescription(description string) (Category, bool) {
	for k, v := range Categories {
		if v.Description == description {
			return k, true
		}
	}
	return -1, false
}
