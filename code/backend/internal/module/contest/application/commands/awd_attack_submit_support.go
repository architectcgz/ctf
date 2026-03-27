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
	acceptedFlags  []string
}

func (s *AWDService) prepareSubmitAttackContext(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*submitAttackContext, error) {
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
	challengeItem, err := s.loadChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureContestChallenge(ctx, contestID, challengeID); err != nil {
		return nil, err
	}

	acceptedFlags, err := s.resolveAcceptedRoundFlags(ctx, contestID, round, req.VictimTeamID, challengeItem, now)
	if err != nil {
		return nil, err
	}
	return &submitAttackContext{
		attackerTeamID: attackerTeamID,
		round:          round,
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
