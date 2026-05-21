package commands

import (
	"context"
	"time"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/internal/platform/randomstring"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"go.uber.org/zap"
)

type rotatedRoundFlag struct {
	previous string
	next     string
}

func (s *AWDService) rotateCurrentRoundFlag(
	ctx context.Context,
	attackContext *submitAttackContext,
	victimTeamID, serviceID int64,
) (*rotatedRoundFlag, bool, error) {
	if attackContext == nil || attackContext.contest == nil || attackContext.round == nil || attackContext.runtimeService == nil {
		return nil, false, apperror.ErrInternal.WithMessage("awd flag rotation context unavailable")
	}
	if s.stateStore == nil {
		return nil, false, apperror.ErrInternal.WithMessage("awd round state store unavailable")
	}

	nextFlag, err := s.buildRotatedRoundFlag(victimTeamID, attackContext.awdChallengeID, attackContext.flagPrefix)
	if err != nil {
		return nil, false, err
	}
	if nextFlag == attackContext.currentFlag {
		return nil, false, apperror.ErrInternal.WithMessage("awd rotated flag unexpectedly unchanged")
	}

	ttl := s.currentRoundTTL(attackContext.contest, attackContext.round, time.Now().UTC())
	claimed, err := s.stateStore.ReplaceAWDRoundFlagIfMatch(
		ctx,
		attackContext.contest.ID,
		attackContext.round.ID,
		victimTeamID,
		serviceID,
		attackContext.currentFlag,
		nextFlag,
		ttl,
	)
	if err != nil {
		return nil, false, apperror.ErrInternal.WithCause(err)
	}
	if !claimed {
		return nil, false, nil
	}

	assignment := contestports.AWDFlagAssignment{
		ServiceID:      serviceID,
		TeamID:         victimTeamID,
		AWDChallengeID: attackContext.awdChallengeID,
		Flag:           nextFlag,
	}
	if err := s.injectRoundFlagAssignment(ctx, attackContext.contest, attackContext.round, assignment, attackContext.currentFlag, ttl); err != nil {
		return nil, false, err
	}

	return &rotatedRoundFlag{
		previous: attackContext.currentFlag,
		next:     nextFlag,
	}, true, nil
}

func (s *AWDService) restoreRotatedRoundFlag(ctx context.Context, attackContext *submitAttackContext, rotation *rotatedRoundFlag) {
	if attackContext == nil || attackContext.contest == nil || attackContext.round == nil || rotation == nil {
		return
	}
	ttl := s.currentRoundTTL(attackContext.contest, attackContext.round, time.Now().UTC())
	assignment := contestports.AWDFlagAssignment{
		ServiceID:      attackContext.runtimeService.ID,
		TeamID:         attackContext.victimTeamID,
		AWDChallengeID: attackContext.awdChallengeID,
		Flag:           rotation.previous,
	}
	if err := s.stateStore.SetAWDRoundFlag(
		ctx,
		attackContext.contest.ID,
		attackContext.round.ID,
		attackContext.victimTeamID,
		attackContext.runtimeService.ID,
		rotation.previous,
		ttl,
	); err != nil && s.log != nil {
		s.log.Warn("restore_rotated_awd_flag_cache_failed",
			zap.Int64("contest_id", attackContext.contest.ID),
			zap.Int64("round_id", attackContext.round.ID),
			zap.Int64("service_id", attackContext.runtimeService.ID),
			zap.Int64("team_id", attackContext.victimTeamID),
			zap.Error(err),
		)
	}
	if err := s.injectFlagAssignments(ctx, attackContext.contest, attackContext.round, []contestports.AWDFlagAssignment{assignment}); err != nil && s.log != nil {
		s.log.Warn("restore_rotated_awd_flag_injection_failed",
			zap.Int64("contest_id", attackContext.contest.ID),
			zap.Int64("round_id", attackContext.round.ID),
			zap.Int64("service_id", attackContext.runtimeService.ID),
			zap.Int64("team_id", attackContext.victimTeamID),
			zap.Error(err),
		)
	}
}

func (s *AWDService) injectRoundFlagAssignment(
	ctx context.Context,
	contest *contestentity.Contest,
	round *contestentity.AWDRound,
	assignment contestports.AWDFlagAssignment,
	rollbackFlag string,
	ttl time.Duration,
) error {
	if err := s.injectFlagAssignments(ctx, contest, round, []contestports.AWDFlagAssignment{assignment}); err != nil {
		if rollbackFlag != "" && s.stateStore != nil {
			rollbackErr := s.stateStore.SetAWDRoundFlag(
				ctx,
				contest.ID,
				round.ID,
				assignment.TeamID,
				assignment.ServiceID,
				rollbackFlag,
				ttl,
			)
			if rollbackErr != nil && s.log != nil {
				s.log.Warn("rollback_rotated_awd_flag_cache_failed",
					zap.Int64("contest_id", contest.ID),
					zap.Int64("round_id", round.ID),
					zap.Int64("service_id", assignment.ServiceID),
					zap.Int64("team_id", assignment.TeamID),
					zap.Error(rollbackErr),
				)
			}
			if injectorRollbackErr := s.injectFlagAssignments(ctx, contest, round, []contestports.AWDFlagAssignment{{
				ServiceID:      assignment.ServiceID,
				TeamID:         assignment.TeamID,
				AWDChallengeID: assignment.AWDChallengeID,
				Flag:           rollbackFlag,
			}}); injectorRollbackErr != nil && s.log != nil {
				s.log.Warn("rollback_rotated_awd_flag_injection_failed",
					zap.Int64("contest_id", contest.ID),
					zap.Int64("round_id", round.ID),
					zap.Int64("service_id", assignment.ServiceID),
					zap.Int64("team_id", assignment.TeamID),
					zap.Error(injectorRollbackErr),
				)
			}
		}
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *AWDService) injectFlagAssignments(
	ctx context.Context,
	contest *contestentity.Contest,
	round *contestentity.AWDRound,
	assignments []contestports.AWDFlagAssignment,
) error {
	if s == nil || s.flagInjector == nil || len(assignments) == 0 {
		return nil
	}
	return s.flagInjector.InjectRoundFlags(ctx, contest, round, assignments)
}

func (s *AWDService) buildRotatedRoundFlag(teamID, challengeID int64, flagPrefix string) (string, error) {
	if teamID <= 0 || challengeID <= 0 || s == nil || s.flagSecret == "" {
		return "", apperror.ErrInternal.WithMessage("awd flag rotation secret unavailable")
	}

	nonce, err := randomstring.Generate()
	if err != nil {
		return "", apperror.ErrInternal.WithCause(err)
	}
	return flagcrypto.GenerateDynamicFlag(teamID, challengeID, s.flagSecret, nonce, flagPrefix), nil
}

func (s *AWDService) currentRoundTTL(contest *contestentity.Contest, round *contestentity.AWDRound, now time.Time) time.Duration {
	if contest == nil || round == nil {
		return 0
	}
	roundEnd := contestdomain.ContestEffectiveEndTime(contest)
	if round.StartedAt != nil {
		candidate := round.StartedAt.Add(s.awdConfig.RoundInterval).Add(contestdomain.ContestPausedDuration(contest))
		if candidate.Before(roundEnd) {
			roundEnd = candidate
		}
	}
	ttl := roundEnd.Sub(now)
	if ttl <= 0 {
		return time.Second
	}
	return ttl
}
