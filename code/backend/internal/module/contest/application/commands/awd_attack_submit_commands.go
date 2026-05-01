package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func (s *AWDService) SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error) {
	attackContext, err := s.prepareSubmitAttackContext(ctx, userID, contestID, serviceID, req)
	if err != nil {
		return nil, err
	}

	return s.createAttackLog(ctx, contestID, attackContext.round.ID, CreateAttackLogInput{
		AttackerTeamID: attackContext.attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ServiceID:      serviceID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      validateSubmittedAttackFlag(req.Flag, attackContext.acceptedFlags),
	}, model.AWDAttackSourceSubmission, &userID)
}
