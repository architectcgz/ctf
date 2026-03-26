package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
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
	now := time.Now()
	if err := s.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
		if err := txRepo.CreateAttackLog(ctx, logRecord); err != nil {
			return err
		}
		if req.IsSuccess {
			if err := txRepo.ApplyAttackImpactToVictimService(ctx, round.ID, req.VictimTeamID, req.ChallengeID, scoreGained, now); err != nil {
				return err
			}
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	}); err != nil {
		return nil, err
	}
	if err := s.repo.RebuildContestScoreboardCache(ctx, s.redis, contestID); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(ctx, s.redis, contestID, roundID, currentRoundID, req.VictimTeamID, req.ChallengeID, model.AWDServiceStatusCompromised); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestdomain.AWDAttackLogRespFromModel(logRecord, teams[req.AttackerTeamID].Name, teams[req.VictimTeamID].Name), nil
}
