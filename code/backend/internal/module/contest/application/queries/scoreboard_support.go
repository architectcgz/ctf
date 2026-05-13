package queries

import (
	"ctf-platform/internal/model"
)

func teamName(team *model.Team) string {
	if team == nil {
		return ""
	}
	return team.Name
}
