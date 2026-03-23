package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/pkg/errcode"
)

type SubmissionService struct {
	contestRepo       Repository
	repo              ContestSubmissionRepository
	redis             *redislib.Client
	flagValidator     challengecontracts.FlagValidator
	teamRepo          ContestTeamFinder
	scoreboardService *ScoreboardService
	cfg               *config.Config
}

func NewSubmissionService(contestRepo Repository, repo ContestSubmissionRepository, redis *redislib.Client, flagValidator challengecontracts.FlagValidator, teamRepo ContestTeamFinder, scoreboardService *ScoreboardService, cfg *config.Config) *SubmissionService {
	return &SubmissionService{
		contestRepo:       contestRepo,
		repo:              repo,
		redis:             redis,
		flagValidator:     flagValidator,
		teamRepo:          teamRepo,
		scoreboardService: scoreboardService,
		cfg:               cfg,
	}
}

func (s *SubmissionService) SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, ErrContestNotFound) {
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
	finalScore := 0

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

	finalScore, err = s.handleCorrectSubmission(ctx, submission, contestChallenge, teamID)
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
		if err := RegistrationStatusError(registration.Status); err != nil {
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

	err = s.repo.WithinTransaction(ctx, func(txRepo ContestSubmissionRepository) error {
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

func (s *SubmissionService) calculateContestScore(contestChallenge model.ContestChallenge, challengeRecord model.Challenge, solveCount int64) int {
	baseScore := s.resolveContestBaseScore(contestChallenge, challengeRecord)
	if s.scoreboardService != nil {
		return s.scoreboardService.CalculateDynamicScoreWithBase(baseScore, solveCount)
	}
	return CalculateDynamicScore(baseScore, s.cfg.Contest.MinScore, s.cfg.Contest.Decay, solveCount)
}

func (s *SubmissionService) resolveContestBaseScore(contestChallenge model.ContestChallenge, challengeRecord model.Challenge) float64 {
	switch {
	case contestChallenge.ContestScore != nil && *contestChallenge.ContestScore > 0:
		return float64(*contestChallenge.ContestScore)
	case contestChallenge.Points > 0:
		return float64(contestChallenge.Points)
	case challengeRecord.Points > 0:
		return float64(challengeRecord.Points)
	default:
		return s.cfg.Contest.BaseScore
	}
}

func mapSubmissionError(err error) error {
	if appErr, ok := err.(*errcode.AppError); ok {
		return appErr
	}
	return errcode.ErrInternal.WithCause(err)
}

func isContestSubmissionUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return pgErr.ConstraintName == "uk_submissions_contest_team_challenge_correct" ||
		pgErr.ConstraintName == "uk_submissions_contest_user_challenge_correct"
}

type contestSubmissionScoreUpdate struct {
	SubmissionID int64
	TeamID       *int64
	OldScore     int
	NewScore     int
}

func buildContestSubmissionScoreUpdates(submissions []model.Submission, firstBloodBy *int64, recalculatedScore, firstBloodBonus int, currentSubmissionID int64) ([]contestSubmissionScoreUpdate, int) {
	firstBloodSubmissionID := int64(0)
	if firstBloodBy != nil {
		for _, solvedSubmission := range submissions {
			if solvedSubmission.TeamID != nil && *solvedSubmission.TeamID == *firstBloodBy {
				firstBloodSubmissionID = solvedSubmission.ID
				break
			}
		}
	}

	updates := make([]contestSubmissionScoreUpdate, 0, len(submissions))
	currentScore := 0
	for _, solvedSubmission := range submissions {
		newScore := recalculatedScore
		if firstBloodSubmissionID > 0 && solvedSubmission.ID == firstBloodSubmissionID {
			newScore += firstBloodBonus
		}
		updates = append(updates, contestSubmissionScoreUpdate{
			SubmissionID: solvedSubmission.ID,
			TeamID:       solvedSubmission.TeamID,
			OldScore:     solvedSubmission.Score,
			NewScore:     newScore,
		})
		if solvedSubmission.ID == currentSubmissionID {
			currentScore = newScore
		}
	}
	return updates, currentScore
}
