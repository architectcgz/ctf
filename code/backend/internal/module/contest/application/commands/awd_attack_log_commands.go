package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error) {
	return s.createAttackLog(ctx, contestID, roundID, req, model.AWDAttackSourceManual)
}

func (s *AWDService) createAttackLog(
	ctx context.Context,
	contestID, roundID int64,
	req *dto.CreateAWDAttackLogReq,
	source string,
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
	if err := s.ensureContestChallenge(ctx, contestID, req.ChallengeID); err != nil {
		return nil, err
	}

	scoreGained := 0
	if req.IsSuccess {
		count, err := s.repo.CountSuccessfulAttacks(ctx, roundID, req.AttackerTeamID, req.VictimTeamID, req.ChallengeID)
		if err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if count == 0 {
			scoreGained = round.AttackScore
		}
	}

	logRecord := &model.AWDAttackLog{
		RoundID:        roundID,
		AttackerTeamID: req.AttackerTeamID,
		VictimTeamID:   req.VictimTeamID,
		ChallengeID:    req.ChallengeID,
		AttackType:     req.AttackType,
		Source:         contestdomain.NormalizeAWDAttackSource(source),
		SubmittedFlag:  req.SubmittedFlag,
		IsSuccess:      req.IsSuccess,
		ScoreGained:    scoreGained,
	}
	if err := s.persistAttackLogAndScores(ctx, contestID, round.ID, req, logRecord); err != nil {
		return nil, err
	}
	return s.buildAttackLogResponse(ctx, contestID, roundID, req, logRecord, teams)
}
