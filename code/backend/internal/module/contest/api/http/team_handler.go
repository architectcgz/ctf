package http

import (
	"context"
	"ctf-platform/internal/dto"
)

type teamCommandService interface {
	CreateTeam(ctx context.Context, contestID, captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error)
	JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*dto.TeamResp, error)
	LeaveTeam(ctx context.Context, contestID, userID, teamID int64) error
	DismissTeam(ctx context.Context, contestID, captainID, teamID int64) error
	KickMember(ctx context.Context, contestID, captainID, teamID, memberUserID int64) error
}

type teamQueryService interface {
	GetTeamInfo(ctx context.Context, teamID int64) (*dto.TeamResp, []*dto.TeamMemberResp, error)
	GetMyTeam(ctx context.Context, contestID, userID int64) (map[string]any, error)
	ListTeams(ctx context.Context, contestID int64) ([]*dto.TeamResp, error)
}

type TeamHandler struct {
	commands teamCommandService
	queries  teamQueryService
}

func NewTeamHandler(commands teamCommandService, queries teamQueryService) *TeamHandler {
	return &TeamHandler{commands: commands, queries: queries}
}
