package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ContestService) UpdateContest(ctx context.Context, id int64, req UpdateContestInput) (*dto.ContestResp, error) {
	contest, err := s.loadContestForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	previousStatus := contest.Status
	previousStatusVersion := contest.StatusVersion
	if err := validateContestUpdateRequest(contest, req); err != nil {
		return nil, err
	}
	if contestdomain.ShouldGateAWDContestStart(contest.Mode, contest.Status, req.Status) {
		if err := ensureAWDReadinessGate(ctx, s.awdRepo, contest.ID, req.ForceOverride, req.OverrideReason); err != nil {
			return nil, err
		}
	}
	if err := applyContestUpdateFields(contest, req); err != nil {
		return nil, err
	}

	if previousStatus != contest.Status {
		contest.UpdatedAt = time.Now().UTC()
		if err := s.updateContestWithManualStatusTransition(ctx, contest, previousStatus, previousStatusVersion); err != nil {
			return nil, err
		}
	} else if err := s.repo.Update(ctx, contest); err != nil {
		s.log.Error("update_contest_failed", zap.Int64("contest_id", id), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("contest_updated", zap.Int64("contest_id", id))
	return contestRespFromModel(contest), nil
}

func (s *ContestService) updateContestWithManualStatusTransition(ctx context.Context, contest *model.Contest, fromStatus string, fromVersion int64) error {
	if contest == nil {
		return errcode.ErrContestNotFound
	}
	if s.transitionRepo == nil {
		s.log.Error("contest_status_transition_repo_missing", zap.Int64("contest_id", contest.ID))
		return errcode.ErrInternal
	}

	result, err := s.transitionRepo.UpdateContestWithStatusTransition(ctx, contest, contestdomain.ContestStatusTransition{
		ContestID:         contest.ID,
		FromStatus:        fromStatus,
		ToStatus:          contest.Status,
		FromStatusVersion: fromVersion,
		Reason:            contestdomain.ContestStatusTransitionReasonManualUpdate,
		OccurredAt:        contest.UpdatedAt.UTC(),
		AppliedBy:         "contest_service",
	})
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		s.log.Error("update_contest_status_transition_failed", zap.Int64("contest_id", contest.ID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}
	if !result.Applied {
		return errcode.New(errcode.ErrConflict.Code, "竞赛状态已变化，请刷新后重试", errcode.ErrConflict.HTTPStatus).
			WithCause(fmt.Errorf("contest %d status changed before manual transition commit", contest.ID))
	}
	contest.StatusVersion = result.StatusVersion

	// Manual status变更也必须走同一套副作用协议；否则 DB 已经进入 frozen/ended，缓存和运行态却还停留在旧阶段。
	if err := s.sideEffects.Run(ctx, result); err != nil {
		if result.RecordID > 0 {
			if markErr := s.transitionRepo.MarkTransitionSideEffectsFailed(ctx, result.RecordID, err); markErr != nil {
				s.log.Error("mark_manual_contest_status_transition_failed",
					zap.Int64("contest_id", contest.ID),
					zap.Int64("transition_record_id", result.RecordID),
					zap.Error(markErr),
				)
			}
		}
		s.log.Error("manual_contest_status_transition_side_effects_failed",
			zap.Int64("contest_id", contest.ID),
			zap.String("from_status", fromStatus),
			zap.String("to_status", contest.Status),
			zap.Error(err),
		)
		return errcode.ErrInternal.WithCause(err)
	}
	if result.RecordID > 0 {
		if err := s.transitionRepo.MarkTransitionSideEffectsSucceeded(ctx, result.RecordID); err != nil {
			s.log.Error("mark_manual_contest_status_transition_succeeded_failed",
				zap.Int64("contest_id", contest.ID),
				zap.Int64("transition_record_id", result.RecordID),
				zap.Error(err),
			)
		}
	}
	return nil
}
