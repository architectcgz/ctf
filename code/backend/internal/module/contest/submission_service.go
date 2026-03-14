package contest

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/config"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
	"ctf-platform/pkg/errcode"
)

type SubmissionService struct {
	db                *gorm.DB
	redis             *redislib.Client
	flagService       *challenge.FlagService
	teamRepo          *TeamRepository
	scoreboardService *ScoreboardService
	cfg               *config.Config
}

func NewSubmissionService(db *gorm.DB, redis *redislib.Client, flagService *challenge.FlagService, teamRepo *TeamRepository, scoreboardService *ScoreboardService, cfg *config.Config) *SubmissionService {
	return &SubmissionService{
		db:                db,
		redis:             redis,
		flagService:       flagService,
		teamRepo:          teamRepo,
		scoreboardService: scoreboardService,
		cfg:               cfg,
	}
}

func (s *SubmissionService) SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	var contest model.Contest
	if err := s.db.WithContext(ctx).First(&contest, contestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	var contestChallenge model.ContestChallenge
	if err := s.db.WithContext(ctx).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
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

	isCorrect, err := s.flagService.ValidateFlag(userID, challengeID, flag, "")
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
		if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
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
	var registration model.ContestRegistration
	if err := s.db.WithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err == nil {
		if err := registrationStatusError(registration.Status); err != nil {
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

func (s *SubmissionService) handleCorrectSubmission(ctx context.Context, submission *model.Submission, contestChallenge model.ContestChallenge, teamID *int64) (int, error) {
	var challengeRecord model.Challenge
	if err := s.db.WithContext(ctx).First(&challengeRecord, submission.ChallengeID).Error; err != nil {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	finalScore := 0
	teamScoreDeltas := make(map[int64]int)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockedChallenge := contestChallenge
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND challenge_id = ?", *submission.ContestID, submission.ChallengeID).
			First(&lockedChallenge).Error; err != nil {
			return err
		}

		solvedQuery := tx.Model(&model.Submission{}).
			Where("contest_id = ? AND challenge_id = ? AND is_correct = ?",
				*submission.ContestID, submission.ChallengeID, true)
		if teamID != nil {
			solvedQuery = solvedQuery.Where("team_id = ?", *teamID)
		} else {
			solvedQuery = solvedQuery.Where("team_id IS NULL AND user_id = ?", submission.UserID)
		}

		var count int64
		if err := solvedQuery.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.ErrContestChallengeSolved
		}

		if teamID != nil && lockedChallenge.FirstBloodBy == nil {
			if err := tx.Model(&model.ContestChallenge{}).
				Where("contest_id = ? AND challenge_id = ?", *submission.ContestID, submission.ChallengeID).
				Update("first_blood_by", *teamID).Error; err != nil {
				return err
			}
			lockedChallenge.FirstBloodBy = teamID
		}

		submission.IsCorrect = true
		submission.Score = 0
		if err := tx.Create(submission).Error; err != nil {
			if isContestSubmissionUniqueViolation(err) {
				return errcode.ErrContestChallengeSolved
			}
			return err
		}

		var solvedSubmissions []model.Submission
		if err := tx.
			Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", *submission.ContestID, submission.ChallengeID, true).
			Order("submitted_at ASC, id ASC").
			Find(&solvedSubmissions).Error; err != nil {
			return err
		}

		recalculatedScore := s.calculateContestScore(lockedChallenge, challengeRecord, int64(len(solvedSubmissions)))
		firstBloodBonus := int(math.Round(float64(recalculatedScore) * s.cfg.Contest.FirstBloodBonus))
		scoreUpdates, currentScore := buildContestSubmissionScoreUpdates(solvedSubmissions, lockedChallenge.FirstBloodBy, recalculatedScore, firstBloodBonus, submission.ID)
		for _, update := range scoreUpdates {
			if update.NewScore == update.OldScore {
				continue
			}
			if err := tx.Model(&model.Submission{}).
				Where("id = ?", update.SubmissionID).
				Update("score", update.NewScore).Error; err != nil {
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
			updates := map[string]any{
				"total_score": gorm.Expr("total_score + ?", delta),
			}
			if teamID != nil && affectedTeamID == *teamID {
				updates["last_solve_at"] = submission.SubmittedAt
			}
			if err := tx.Model(&model.Team{}).
				Where("id = ?", affectedTeamID).
				Updates(updates).Error; err != nil {
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
	return calculateDynamicScore(baseScore, s.cfg.Contest.MinScore, s.cfg.Contest.Decay, solveCount)
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
