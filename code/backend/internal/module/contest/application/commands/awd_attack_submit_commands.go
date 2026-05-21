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

	matchedFlag, isSuccess := matchSubmittedAttackFlag(req.Flag, attackContext.acceptedFlags)
	var rotation *rotatedRoundFlag
	if isSuccess && matchedFlag == attackContext.currentFlag {
		var claimed bool
		rotation, claimed, err = s.rotateCurrentRoundFlag(ctx, attackContext, req.VictimTeamID, serviceID)
		if err != nil {
			return nil, err
		}
		if !claimed {
			isSuccess = false
		}
	}

	resp, err := s.createAttackLog(ctx, contestID, attackContext.round.ID, CreateAttackLogInput{
		AttackerTeamID: attackContext.attackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ServiceID:      serviceID,
		AttackType:     contestentity.AWDAttackTypeFlagCapture,
		SubmittedFlag:  req.Flag,
		IsSuccess:      isSuccess,
	}, contestentity.AWDAttackSourceSubmission, &userID)
	if err != nil {
		if rotation != nil {
			s.restoreRotatedRoundFlag(ctx, attackContext, rotation)
		}
		return nil, err
	}
	return resp, nil
}
