package commands

import (
	"context"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) CreateRound(ctx context.Context, contestID int64, req CreateAWDRoundInput) (*AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := ensureAWDReadinessGate(ctx, s.repo, contestID, req.ForceOverride, req.OverrideReason); err != nil {
		return nil, err
	}

	round := &contestentity.AWDRound{
		ContestID:    contestID,
		RoundNumber:  req.RoundNumber,
		Status:       contestentity.AWDRoundStatusPending,
		AttackScore:  contestdomain.AWDDefaultRoundAttackScore,
		DefenseScore: contestdomain.AWDDefaultRoundDefenseScore,
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
	if round.AttackScore < 0 || round.AttackScore > contestdomain.AWDMaxRoundAttackScore {
		return nil, errcode.ErrInvalidParams
	}
	if round.DefenseScore < 0 || round.DefenseScore > contestdomain.AWDMaxRoundDefenseScore {
		return nil, errcode.ErrInvalidParams
	}
	if err := s.repo.CreateRound(ctx, round); err != nil {
		if contestdomain.IsUniqueConstraintError(err) {
			return nil, errcode.ErrConflict
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	roundResp := contestResponseMapperInst.ToAWDRoundRespBasePtr(round)
	if roundResp == nil {
		return nil, nil
	}
	return &AWDRoundResp{
		ID:           roundResp.ID,
		ContestID:    roundResp.ContestID,
		RoundNumber:  roundResp.RoundNumber,
		Status:       roundResp.Status,
		StartedAt:    roundResp.StartedAt,
		EndedAt:      roundResp.EndedAt,
		AttackScore:  roundResp.AttackScore,
		DefenseScore: roundResp.DefenseScore,
		CreatedAt:    roundResp.CreatedAt,
		UpdatedAt:    roundResp.UpdatedAt,
	}, nil
}
