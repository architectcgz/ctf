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
		if registration.Status != "" && registration.Status != "approved" {
			return nil, errcode.ErrRegistrationNotApproved
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

	baseScore := contestChallenge.Points
	if contestChallenge.ContestScore != nil {
		baseScore = *contestChallenge.ContestScore
	}
	finalScore := baseScore

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Submission{}).
			Where("contest_id = ? AND user_id = ? AND challenge_id = ? AND is_correct = ?",
				*submission.ContestID, submission.UserID, submission.ChallengeID, true).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errcode.ErrContestChallengeSolved
		}

		if teamID != nil {
			result := tx.Model(&model.ContestChallenge{}).
				Where("contest_id = ? AND challenge_id = ? AND first_blood_by IS NULL", *submission.ContestID, submission.ChallengeID).
				Update("first_blood_by", *teamID)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				finalScore += int(math.Round(float64(baseScore) * s.cfg.Contest.FirstBloodBonus))
			}

			if err := tx.Model(&model.Team{}).
				Where("id = ?", *teamID).
				Updates(map[string]any{
					"total_score":   gorm.Expr("total_score + ?", finalScore),
					"last_solve_at": submission.SubmittedAt,
				}).Error; err != nil {
				return err
			}
		}

		submission.IsCorrect = true
		submission.Score = finalScore
		if err := tx.Create(submission).Error; err != nil {
			if isContestSubmissionUniqueViolation(err) {
				return errcode.ErrContestChallengeSolved
			}
			return err
		}
		return nil
	})
	if err != nil {
		return 0, mapSubmissionError(err)
	}

	if teamID != nil && s.scoreboardService != nil {
		if err := s.scoreboardService.UpdateScore(ctx, *submission.ContestID, *teamID, float64(finalScore)); err != nil {
			return 0, err
		}
	}
	return finalScore, nil
}

func mapSubmissionError(err error) error {
	if appErr, ok := err.(*errcode.AppError); ok {
		return appErr
	}
	return errcode.ErrInternal.WithCause(err)
}

func isContestSubmissionUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uk_submissions_contest_user_challenge_correct"
}
