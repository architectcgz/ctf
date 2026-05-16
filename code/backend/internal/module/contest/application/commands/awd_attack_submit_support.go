package commands

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type submitAttackContext struct {
	attackerTeamID int64
	round          *model.AWDRound
	runtimeService *model.ContestAWDService
	awdChallengeID int64
	flagPrefix     string
	acceptedFlags  []string
}

func (s *AWDService) prepareSubmitAttackContext(ctx context.Context, userID, contestID, serviceID int64, req SubmitAttackInput) (*submitAttackContext, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if contestdomain.ContestHasEndedAt(contest, now) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}

	attackerTeamID, err := s.resolveUserTeamID(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}
	round, err := s.resolveCurrentRoundForContest(ctx, contest)
	if err != nil {
		return nil, err
	}
	runtimeService, err := s.resolveContestRuntimeService(ctx, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	snapshot, _ := model.DecodeContestAWDServiceSnapshot(runtimeService.ServiceSnapshot)
	flagPrefix := resolveSubmitAttackFlagPrefix(snapshot)

	acceptedFlags, err := s.resolveAcceptedRoundFlags(ctx, contest, contestID, round, req.VictimTeamID, runtimeService.AWDChallengeID, flagPrefix, runtimeService.ID, now)
	if err != nil {
		return nil, err
	}
	return &submitAttackContext{
		attackerTeamID: attackerTeamID,
		round:          round,
		runtimeService: runtimeService,
		awdChallengeID: runtimeService.AWDChallengeID,
		flagPrefix:     flagPrefix,
		acceptedFlags:  acceptedFlags,
	}, nil
}

func resolveSubmitAttackFlagPrefix(snapshot model.ContestAWDServiceSnapshot) string {
	if snapshot.FlagConfig != nil {
		if value, ok := snapshot.FlagConfig["flag_prefix"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "flag"
}

func validateSubmittedAttackFlag(submittedFlag string, acceptedFlags []string) bool {
	for _, candidate := range acceptedFlags {
		if crypto.ValidateFlag(submittedFlag, candidate) {
			return true
		}
	}
	return false
}
