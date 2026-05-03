package commands

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

func (s *Service) SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	challengeItem, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotFound
		}
		s.logger.Error("查询靶场失败", zap.Int64("challenge_id", challengeID), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if challengeItem.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	alreadySolved := false
	if _, err := s.repo.FindCorrectSubmission(ctx, userID, challengeID); err == nil {
		alreadySolved = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if alreadySolved && challengeItem.FlagType == model.FlagTypeManualReview {
		return nil, errcode.ErrAlreadySolved
	}

	rateLimitKey := fmt.Sprintf("%s:submit:%d:%d", s.config.RateLimit.RedisKeyPrefix, userID, challengeID)
	count, err := s.redis.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		s.logger.Error("提交限流失败", zap.String("key", rateLimitKey), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if count == 1 {
		_ = s.redis.Expire(ctx, rateLimitKey, s.config.RateLimit.FlagSubmit.Window).Err()
	}
	if count > int64(s.config.RateLimit.FlagSubmit.Limit) {
		return nil, errcode.ErrSubmitTooFrequent
	}

	submission := &model.Submission{
		UserID:       userID,
		ChallengeID:  challengeID,
		Flag:         "",
		IsCorrect:    false,
		ReviewStatus: model.SubmissionReviewStatusNotRequired,
		SubmittedAt:  time.Now(),
		UpdatedAt:    time.Now(),
	}
	status := dto.SubmissionStatusIncorrect
	submissionPersisted := false

	if challengeItem.FlagType == model.FlagTypeManualReview {
		submission.Flag = flag
		submission.ReviewStatus = model.SubmissionReviewStatusPending
		status = dto.SubmissionStatusPendingReview
	} else {
		isCorrect, err := s.validateSubmittedFlag(ctx, userID, challengeItem, flag)
		if err != nil {
			return nil, err
		}
		submission.IsCorrect = isCorrect
		if isCorrect {
			status = dto.SubmissionStatusCorrect
			if alreadySolved {
				auditlog.MarkSkip(ctx)
				return &dto.SubmissionResp{
					IsCorrect:   true,
					Status:      status,
					SubmittedAt: submission.SubmittedAt,
				}, nil
			}
		}
	}

	if !submissionPersisted {
		if err := s.repo.CreateSubmission(ctx, submission); err != nil {
			if submission.IsCorrect && s.repo.IsUniqueViolation(err) {
				return nil, errcode.ErrAlreadySolved
			}
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	if submission.IsCorrect && !alreadySolved {
		cacheKey := constants.UserProgressKey(userID)
		if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
			s.logger.Warn("删除进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		}
		s.publishWeakEvent(ctx, platformevents.Event{
			Name: practicecontracts.EventFlagAccepted,
			Payload: practicecontracts.FlagAcceptedEvent{
				UserID:      userID,
				ChallengeID: challengeID,
				Dimension:   challengeItem.Category,
				Points:      challengeItem.Points,
				OccurredAt:  submission.SubmittedAt,
			},
		})
	}

	var instanceShutdownAt *time.Time
	if submission.IsCorrect && !alreadySolved {
		instanceShutdownAt = s.applySolveGracePeriod(ctx, userID, challengeItem, submission.SubmittedAt)
	}

	resp := &dto.SubmissionResp{
		IsCorrect:          submission.IsCorrect,
		Status:             status,
		SubmittedAt:        submission.SubmittedAt,
		InstanceShutdownAt: instanceShutdownAt,
	}
	if submission.IsCorrect && !alreadySolved {
		resp.Points = challengeItem.Points
		if s.scoreService != nil {
			s.triggerScoreUpdate(userID)
		}
	}

	return resp, nil
}

func (s *Service) applySolveGracePeriod(ctx context.Context, userID int64, challengeItem *model.Challenge, solvedAt time.Time) *time.Time {
	if s == nil || s.instanceRepo == nil || challengeItem == nil {
		return nil
	}

	gracePeriod := s.config.Container.SolveGracePeriod
	if gracePeriod <= 0 {
		return nil
	}

	instance, err := s.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeItem.ID)
	if err != nil {
		s.logger.Warn("查询解题后实例失败", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeItem.ID), zap.Error(err))
		return nil
	}
	if instance == nil || instance.ShareScope != model.InstanceSharingPerUser {
		return nil
	}

	shutdownAt := instance.ExpiresAt
	graceExpiry := solvedAt.Add(gracePeriod)
	if shutdownAt.After(graceExpiry) {
		shutdownAt = graceExpiry
		if err := s.instanceRepo.RefreshInstanceExpiry(ctx, instance.ID, shutdownAt); err != nil {
			s.logger.Warn("收缩解题后实例生命周期失败", zap.Int64("instance_id", instance.ID), zap.Error(err))
			return nil
		}
	}

	return &shutdownAt
}

func formatSolveGracePeriod(delay time.Duration) string {
	if delay <= 0 || delay < time.Minute {
		return "1 分钟内"
	}
	if delay%time.Hour == 0 {
		return fmt.Sprintf("%d 小时", int(delay/time.Hour))
	}

	minutes := int(delay.Round(time.Minute) / time.Minute)
	if minutes <= 1 {
		return "1 分钟"
	}
	return fmt.Sprintf("%d 分钟", minutes)
}

func (s *Service) buildInstanceFlag(subjectID, challengeID int64, chal *model.Challenge) (string, string, error) {
	switch chal.FlagType {
	case model.FlagTypeDynamic:
		nonce, err := crypto.GenerateNonce()
		if err != nil {
			return "", "", errcode.ErrInternal.WithCause(err)
		}
		if s.config.Container.FlagGlobalSecret == "" {
			return "", "", errcode.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		flag := crypto.GenerateDynamicFlag(subjectID, challengeID, s.config.Container.FlagGlobalSecret, nonce, chal.FlagPrefix)
		return flag, nonce, nil
	case model.FlagTypeStatic:
		return chal.FlagHash, "", nil
	default:
		return "", "", nil
	}
}

func (s *Service) validateSubmittedFlag(ctx context.Context, userID int64, challengeItem *model.Challenge, flag string) (bool, error) {
	switch challengeItem.FlagType {
	case model.FlagTypeStatic:
		inputHash := crypto.HashStaticFlag(flag, challengeItem.FlagSalt)
		return crypto.ValidateFlag(inputHash, challengeItem.FlagHash), nil
	case model.FlagTypeRegex:
		return regexp.MatchString(challengeItem.FlagRegex, flag)
	case model.FlagTypeManualReview:
		return false, nil
	case model.FlagTypeDynamic:
	default:
		return false, errcode.ErrInvalidParams.WithCause(fmt.Errorf("unsupported flag type %s", challengeItem.FlagType))
	}

	instance, err := s.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeItem.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, errcode.ErrInternal.WithCause(err)
	}
	if instance == nil || instance.Nonce == "" || s.config.Container.FlagGlobalSecret == "" {
		return false, nil
	}

	expectedFlag := crypto.GenerateDynamicFlag(userID, challengeItem.ID, s.config.Container.FlagGlobalSecret, instance.Nonce, challengeItem.FlagPrefix)
	return crypto.ValidateFlag(flag, expectedFlag), nil
}
