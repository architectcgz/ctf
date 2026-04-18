package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type submitAttackContext struct {
	attackerTeamID int64
	round          *model.AWDRound
	runtimeService *model.ContestAWDService
	challenge      *model.Challenge
	acceptedFlags  []string
}

func (s *AWDService) prepareSubmitAttackContext(ctx context.Context, userID, contestID, serviceID int64, req *dto.SubmitAWDAttackReq) (*submitAttackContext, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
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
	challengeItem, err := s.loadChallenge(ctx, runtimeService.ChallengeID)
	if err != nil {
		return nil, err
	}

	acceptedFlags, err := s.resolveAcceptedRoundFlags(ctx, contestID, round, req.VictimTeamID, challengeItem, runtimeService.ID, now)
	if err != nil {
		return nil, err
	}
	return &submitAttackContext{
		attackerTeamID: attackerTeamID,
		round:          round,
		runtimeService: runtimeService,
		challenge:      challengeItem,
		acceptedFlags:  acceptedFlags,
	}, nil
}

func validateSubmittedAttackFlag(submittedFlag string, acceptedFlags []string) bool {
	for _, candidate := range acceptedFlags {
		if crypto.ValidateFlag(submittedFlag, candidate) {
			return true
		}
	}
	return false
}
