package queries

import (
	contestentity "ctf-platform/internal/module/contest/entity"
)

func teamName(team *contestentity.Team) string {
	if team == nil {
		return ""
	}
	return team.Name
}
