package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) GetMyProgress(ctx context.Context, contestID, userID int64) (*dto.ContestMyProgressResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamID, err := s.resolveUserTeamID(ctx, contestID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListSolvedProgress(ctx, contestID, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := &dto.ContestMyProgressResp{
		ContestID: contestID,
		TeamID:    teamID,
		Solved:    make([]*dto.ContestSolvedProgressItem, 0, len(rows)),
	}
	for _, row := range rows {
		result.Solved = append(result.Solved, &dto.ContestSolvedProgressItem{
			ContestChallengeID: row.ContestChallengeID,
			SolvedAt:           row.SolvedAt,
			PointsEarned:       row.PointsEarned,
		})
	}
	return result, nil
}

func (s *ParticipationService) resolveUserTeamID(ctx context.Context, contestID, userID int64) (*int64, error) {
	if registration, err := s.repo.FindRegistration(ctx, contestID, userID); err == nil {
		return registration.TeamID, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err == nil && team != nil && team.ID > 0 {
		return &team.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return nil, nil
}
