package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error) {
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
	isSuccess := false
	for _, candidate := range acceptedFlags {
		if crypto.ValidateFlag(req.Flag, candidate) {
			isSuccess = true
			break
		}
	}

	return s.createAttackLog(ctx, contestID, round.ID, &dto.CreateAWDAttackLogReq{
		AttackerTeamID: attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ChallengeID:    challengeID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      isSuccess,
	}, model.AWDAttackSourceSubmission)
}
