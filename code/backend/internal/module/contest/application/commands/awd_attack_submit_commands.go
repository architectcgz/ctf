package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func (s *AWDService) SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error) {
	attackContext, err := s.prepareSubmitAttackContext(ctx, userID, contestID, challengeID, req)
	if err != nil {
		return nil, err
	}

	return s.createAttackLog(ctx, contestID, attackContext.round.ID, &dto.CreateAWDAttackLogReq{
		AttackerTeamID: attackContext.attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ChallengeID:    challengeID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      validateSubmittedAttackFlag(req.Flag, attackContext.acceptedFlags),
	}, model.AWDAttackSourceSubmission)
}
