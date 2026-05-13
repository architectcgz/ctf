package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) CreateTeam(ctx context.Context, contestID, captainID int64, req CreateTeamInput) (*dto.TeamResp, error) {
	if err := s.ensureTeamJoinableContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := s.ensureApprovedRegistration(ctx, contestID, captainID); err != nil {
		return nil, err
	}

	existingTeam, err := s.teamRepo.FindUserTeamInContest(ctx, captainID, contestID)
	if err != nil {
		if !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	} else if existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	team, err := s.createTeamWithInviteRetries(ctx, contestID, captainID, req.Name, resolveCreateTeamMaxMembers(req))
	if err != nil {
		return nil, err
	}
	return teamRespFromModel(team, 1), nil
}
