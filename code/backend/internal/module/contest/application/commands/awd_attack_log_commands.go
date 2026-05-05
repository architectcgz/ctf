package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
	"gorm.io/gorm"
)

func (s *AWDService) CreateAttackLog(ctx context.Context, contestID, roundID int64, req CreateAttackLogInput) (*dto.AWDAttackLogResp, error) {
	return s.createAttackLog(ctx, contestID, roundID, req, model.AWDAttackSourceManual, nil)
}

func (s *AWDService) createAttackLog(
	ctx context.Context,
	contestID, roundID int64,
	req CreateAttackLogInput,
	source string,
	submittedByUserID *int64,
) (*dto.AWDAttackLogResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	if req.AttackerTeamID == req.VictimTeamID {
		return nil, errcode.ErrInvalidParams
	}
	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if teams[req.AttackerTeamID] == nil || teams[req.VictimTeamID] == nil {
		return nil, errcode.ErrNotFound
	}
	runtimeService, err := s.resolveContestRuntimeService(ctx, contestID, req.ServiceID)
	if err != nil {
		return nil, err
	}

	scoreGained := 0
	if req.IsSuccess {
		count, err := s.repo.CountSuccessfulAttacks(ctx, roundID, req.AttackerTeamID, req.VictimTeamID, runtimeService.ID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if count == 0 {
			scoreGained = round.AttackScore
		}
	}

	logRecord := &model.AWDAttackLog{
		RoundID:           roundID,
		AttackerTeamID:    req.AttackerTeamID,
		VictimTeamID:      req.VictimTeamID,
		ServiceID:         runtimeService.ID,
		AWDChallengeID:    runtimeService.AWDChallengeID,
		AttackType:        req.AttackType,
		Source:            contestdomain.NormalizeAWDAttackSource(source),
		SubmittedFlag:     req.SubmittedFlag,
		SubmittedByUserID: submittedByUserID,
		IsSuccess:         req.IsSuccess,
		ScoreGained:       scoreGained,
	}
	if err := s.persistAttackLogAndScores(ctx, contestID, round.ID, req, logRecord); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	resp, err := s.buildAttackLogResponse(ctx, contestID, roundID, req, logRecord, teams)
	if err != nil {
		return nil, err
	}
	if submittedByUserID != nil && logRecord.IsSuccess && logRecord.ScoreGained > 0 {
		snapshot, _ := model.DecodeContestAWDServiceSnapshot(runtimeService.ServiceSnapshot)
		s.publishWeakEvent(ctx, events.Event{
			Name: contestcontracts.EventAWDAttackAccepted,
			Payload: contestcontracts.AWDAttackAcceptedEvent{
				UserID:         *submittedByUserID,
				ContestID:      contestID,
				AWDChallengeID: runtimeService.AWDChallengeID,
				Dimension:      snapshot.Category,
				OccurredAt:     logRecord.CreatedAt,
			},
		})
	}
	return resp, nil
}
