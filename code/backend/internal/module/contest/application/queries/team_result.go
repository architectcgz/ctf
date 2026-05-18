package queries

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
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

func teamResultFromModel(team *contestentity.Team, memberCount int) *TeamResult {
	resp := contestQueryResponseMapperInst.ToTeamResultBasePtr(team)
	if resp == nil {
		return nil
	}
	resp.MemberCount = memberCount
	return resp
}
