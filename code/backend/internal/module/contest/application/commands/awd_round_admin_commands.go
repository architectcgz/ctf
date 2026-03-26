package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round := &model.AWDRound{
		ContestID:    contestID,
		RoundNumber:  req.RoundNumber,
		Status:       model.AWDRoundStatusPending,
		AttackScore:  50,
		DefenseScore: 50,
	}
	if req.Status != nil && *req.Status != "" {
		round.Status = *req.Status
	}
	if req.AttackScore != nil {
		round.AttackScore = *req.AttackScore
	}
	if req.DefenseScore != nil {
		round.DefenseScore = *req.DefenseScore
	}
	if err := s.repo.CreateRound(ctx, round); err != nil {
		if contestdomain.IsUniqueConstraintError(err) {
			return nil, errcode.ErrConflict
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contestdomain.AWDRoundRespFromModel(round), nil
}
