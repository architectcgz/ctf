package http

import (
	"context"

	"ctf-platform/internal/dto"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestqry "ctf-platform/internal/module/contest/application/queries"
)

type teamCommandService interface {
	CreateTeam(ctx context.Context, contestID, captainID int64, req contestcmd.CreateTeamInput) (*dto.TeamResp, error)
	JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*dto.TeamResp, error)
	LeaveTeam(ctx context.Context, contestID, userID, teamID int64) error
	DismissTeam(ctx context.Context, contestID, captainID, teamID int64) error
	KickMember(ctx context.Context, contestID, captainID, teamID, memberUserID int64) error
}

type teamQueryService interface {
	GetTeamInfo(ctx context.Context, teamID int64) (*contestqry.TeamResult, []*contestqry.TeamMemberResult, error)
	GetMyTeam(ctx context.Context, contestID, userID int64) (*contestqry.MyTeamResult, error)
	ListTeams(ctx context.Context, contestID int64) ([]*contestqry.TeamResult, error)
}

type TeamHandler struct {
	commands teamCommandService
	queries  teamQueryService
}

func NewTeamHandler(commands teamCommandService, queries teamQueryService) *TeamHandler {
	return &TeamHandler{commands: commands, queries: queries}
}
