package commands

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *AWDService) SubmitAttack(ctx context.Context, userID, contestID, serviceID int64, req SubmitAttackInput) (*AWDAttackLogResp, error) {
	attackContext, err := s.prepareSubmitAttackContext(ctx, userID, contestID, serviceID, req)
	if err != nil {
		return nil, err
	}

	return s.createAttackLog(ctx, contestID, attackContext.round.ID, CreateAttackLogInput{
		AttackerTeamID: attackContext.attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      validateSubmittedAttackFlag(req.Flag, attackContext.acceptedFlags),
	}, contestentity.AWDAttackSourceSubmission, &userID)
}
