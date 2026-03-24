package commands

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}

	teamID, err := s.resolveTeamID(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	contestChallenge, err := s.repo.FindContestChallenge(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotInContest
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rateLimitKey := fmt.Sprintf("contest:submit:rate:%d:%d:%d", userID, contestID, challengeID)
	exists, err := s.redis.Exists(ctx, rateLimitKey).Result()
	if err == nil && exists > 0 {
		return nil, errcode.ErrSubmitTooFrequent
	}

	if s.flagValidator == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("challenge flag validator is nil"))
	}
	isCorrect, err := s.flagValidator.ValidateFlag(userID, challengeID, flag, "")
	if err != nil {
		return nil, err
	}

	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      teamID,
		Flag:        flag,
		IsCorrect:   false,
		Score:       0,
		SubmittedAt: now,
	}

	if !isCorrect {
		_ = s.redis.Set(ctx, rateLimitKey, "1", 5*time.Second).Err()
		if err := s.repo.CreateSubmission(ctx, submission); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return &dto.SubmissionResp{
			IsCorrect:   false,
			Message:     constants.MsgFlagIncorrect,
			SubmittedAt: submission.SubmittedAt,
		}, nil
	}

	finalScore, err := s.handleCorrectSubmission(ctx, submission, contestChallenge, teamID)
	if err != nil {
		return nil, err
	}

	return &dto.SubmissionResp{
		IsCorrect:   true,
		Message:     constants.MsgFlagCorrect,
		Points:      finalScore,
		SubmittedAt: submission.SubmittedAt,
	}, nil
}

func (s *SubmissionService) resolveTeamID(ctx context.Context, userID, contestID int64) (*int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return nil, err
		}
		return registration.TeamID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err == nil && team.ID > 0 {
		return &team.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return nil, errcode.ErrNotRegistered
}

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *model.Submission, contestChallenge *model.ContestChallenge, teamID *int64) (int, error) {
	challengeRecord, err := s.repo.FindChallengeByID(ctx, submission.ChallengeID)
	if err != nil {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	finalScore := 0
	teamScoreDeltas := make(map[int64]int)

	err = s.repo.WithinTransaction(ctx, func(txRepo contestports.ContestSubmissionRepository) error {
		lockedChallenge, err := txRepo.LockContestChallenge(ctx, *submission.ContestID, submission.ChallengeID)
		if err != nil {
			return err
		}

		count, err := txRepo.CountCorrectSubmissions(ctx, *submission.ContestID, submission.ChallengeID, teamID, submission.UserID)
		if err != nil {
			return err
		}
		if count > 0 {
			return errcode.ErrContestChallengeSolved
		}

		if teamID != nil && lockedChallenge.FirstBloodBy == nil {
			if err := txRepo.UpdateFirstBlood(ctx, *submission.ContestID, submission.ChallengeID, *teamID); err != nil {
				return err
			}
			lockedChallenge.FirstBloodBy = teamID
		}

		submission.IsCorrect = true
		submission.Score = 0
		if err := txRepo.CreateSubmission(ctx, submission); err != nil {
			if isContestSubmissionUniqueViolation(err) {
				return errcode.ErrContestChallengeSolved
			}
			return err
		}

		solvedSubmissions, err := txRepo.ListCorrectSubmissions(ctx, *submission.ContestID, submission.ChallengeID)
		if err != nil {
			return err
		}

		recalculatedScore := s.calculateContestScore(*lockedChallenge, *challengeRecord, int64(len(solvedSubmissions)))
		firstBloodBonus := int(math.Round(float64(recalculatedScore) * s.cfg.Contest.FirstBloodBonus))
		scoreUpdates, currentScore := buildContestSubmissionScoreUpdates(solvedSubmissions, lockedChallenge.FirstBloodBy, recalculatedScore, firstBloodBonus, submission.ID)
		for _, update := range scoreUpdates {
			if update.NewScore == update.OldScore {
				continue
			}
			if err := txRepo.UpdateSubmissionScore(ctx, update.SubmissionID, update.NewScore); err != nil {
				return err
			}
			if update.TeamID != nil {
				teamScoreDeltas[*update.TeamID] += update.NewScore - update.OldScore
			}
		}

		for affectedTeamID, delta := range teamScoreDeltas {
			if delta == 0 {
				continue
			}
			var lastSolveAt *time.Time
			if teamID != nil && affectedTeamID == *teamID {
				lastSolveAt = &submission.SubmittedAt
			}
			if err := txRepo.AddTeamScore(ctx, affectedTeamID, delta, lastSolveAt); err != nil {
				return err
			}
		}

		finalScore = currentScore
		return nil
	})
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if submission.ContestID != nil && s.scoreboardService != nil {
		for affectedTeamID, delta := range teamScoreDeltas {
			if delta == 0 {
				continue
			}
			if err := s.scoreboardService.UpdateScore(ctx, *submission.ContestID, affectedTeamID, float64(delta)); err != nil {
				if rebuildErr := s.scoreboardService.RebuildScoreboard(ctx, *submission.ContestID); rebuildErr != nil {
					return 0, rebuildErr
				}
				break
			}
		}
	}
	return finalScore, nil
}
