package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) CreateTeam(ctx context.Context, contestID, captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
	if err := s.ensureTeamJoinableContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := s.ensureApprovedRegistration(ctx, contestID, captainID); err != nil {
		return nil, err
	}

	existingTeam, err := s.teamRepo.FindUserTeamInContest(ctx, captainID, contestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	team, err := s.createTeamWithInviteRetries(ctx, contestID, captainID, req.Name, resolveCreateTeamMaxMembers(req))
	if err != nil {
		return nil, err
	}
	return teamRespFromModel(team, 1), nil
}
