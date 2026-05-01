package queries

import (
	"time"

	"ctf-platform/internal/model"
)

type TeamResult struct {
	ID          int64
	ContestID   int64
	Name        string
	CaptainID   int64
	InviteCode  string
	MaxMembers  int
	MemberCount int
	CreatedAt   time.Time
}

type TeamMemberResult struct {
	UserID   int64
	Username string
	JoinedAt time.Time
}

type MyTeamResult struct {
	ID         int64
	Name       string
	InviteCode string
	CaptainID  int64
	Members    []*TeamMemberResult
}

func teamResultFromModel(team *model.Team, memberCount int) *TeamResult {
	if team == nil {
		return nil
	}
	return &TeamResult{
		ID:          team.ID,
		ContestID:   team.ContestID,
		Name:        team.Name,
		CaptainID:   team.CaptainID,
		InviteCode:  team.InviteCode,
		MaxMembers:  team.MaxMembers,
		MemberCount: memberCount,
		CreatedAt:   team.CreatedAt,
	}
}
