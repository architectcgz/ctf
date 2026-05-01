package commands

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error) {
	contest, err := s.loadContestForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateContestUpdateRequest(contest, req); err != nil {
		return nil, err
	}
	if domain.ShouldGateAWDContestStart(contest.Mode, contest.Status, req.Status) {
		if err := ensureAWDReadinessGate(ctx, s.awdRepo, contest.ID, req.ForceOverride, req.OverrideReason); err != nil {
			return nil, err
		}
	}
	if err := applyContestUpdateFields(contest, req); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, contest); err != nil {
		s.log.Error("update_contest_failed", zap.Int64("contest_id", id), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_updated", zap.Int64("contest_id", id))
	return contestRespFromModel(contest), nil
}
