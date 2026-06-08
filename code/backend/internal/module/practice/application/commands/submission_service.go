package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/platform/randomstring"
	crypto "ctf-platform/internal/shared/flagcrypto"
)

func (s *Service) SubmitFlag(ctx context.Context, userID, challengeID int64, flag string) (*SubmissionResp, error) {
	if s.runtimeSubject == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice runtime subject repository is nil"))
	}
	challengeItem, err := s.runtimeSubject.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		s.logger.Error("查询靶场失败", zap.Int64("challenge_id", challengeID), zap.Error(err))
		return nil, apperror.ErrInternal.WithCause(err)
	}

	if challengeItem.Status != practiceentity.ChallengeStatusPublished {
		return nil, challengecontracts.ErrChallengeNotPublish
	}

	if s.solvedSubmission == nil {
		return nil, apperror.ErrInternal.WithCause(errors.New("practice solved submission repository is nil"))
	}

	alreadySolved := false
	if _, err := s.solvedSubmission.FindCorrectSubmission(ctx, userID, challengeID); err == nil {
		alreadySolved = true
	} else if err != nil && !errors.Is(err, practiceports.ErrPracticeSolvedSubmissionNotFound) {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if alreadySolved && challengeItem.FlagType == practiceentity.FlagTypeManualReview {
		return nil, challengecontracts.ErrAlreadySolved
	}

	if s.rateLimitStore != nil {
		allowed, err := s.rateLimitStore.AllowFlagSubmit(ctx, userID, challengeID, s.config.RateLimit.FlagSubmit.Limit, s.config.RateLimit.FlagSubmit.Window)
		if err != nil {
			s.logger.Error("提交限流失败", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeID), zap.Error(err))
			return nil, apperror.ErrInternal.WithCause(err)
		}
		if !allowed {
			return nil, challengecontracts.ErrSubmitTooFrequent
		}
	}

	now := time.Now().UTC()
	submission := &practiceports.SubmissionRecord{
		UserID:       userID,
		ChallengeID:  challengeID,
		Flag:         "",
		IsCorrect:    false,
		ReviewStatus: practiceports.SubmissionReviewStatusNotRequired,
		SubmittedAt:  now,
		UpdatedAt:    now,
	}
	status := SubmissionStatusIncorrect
	submissionPersisted := false

	if challengeItem.FlagType == practiceentity.FlagTypeManualReview {
		submission.Flag = flag
		submission.ReviewStatus = practiceports.SubmissionReviewStatusPending
		status = SubmissionStatusPendingReview
	} else {
		isCorrect, err := s.validateSubmittedFlag(ctx, userID, challengeItem, flag)
		if err != nil {
			return nil, err
		}
		submission.IsCorrect = isCorrect
		if isCorrect {
			status = SubmissionStatusCorrect
			if alreadySolved {
				auditlog.MarkSkip(ctx)
				return &SubmissionResp{
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
				return nil, challengecontracts.ErrAlreadySolved
			}
			return nil, apperror.ErrInternal.WithCause(err)
		}
	}

	if submission.IsCorrect && !alreadySolved {
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

	resp := &SubmissionResp{
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

func (s *Service) applySolveGracePeriod(ctx context.Context, userID int64, challengeItem *practiceentity.Challenge, solvedAt time.Time) *time.Time {
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
	if instance == nil || instance.ShareScope != instancecontracts.ShareScopePerUser {
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

func (s *Service) buildInstanceFlag(subjectID, challengeID int64, chal *practiceentity.Challenge) (string, string, string, error) {
	switch chal.FlagType {
	case practiceentity.FlagTypeDynamic:
		nonce, err := randomstring.Generate()
		if err != nil {
			return "", "", "", apperror.ErrInternal.WithCause(err)
		}
		keyID := s.activeFlagSecretKeyID()
		secret, ok := s.flagSecretForKeyID(keyID)
		if !ok {
			return "", "", "", apperror.ErrInternal.WithCause(fmt.Errorf("flag global secret is empty"))
		}
		flag := crypto.GenerateDynamicFlag(subjectID, challengeID, secret, nonce, chal.FlagPrefix)
		return flag, nonce, keyID, nil
	case practiceentity.FlagTypeStatic:
		return chal.FlagHash, "", "", nil
	default:
		return "", "", "", nil
	}
}

func (s *Service) validateSubmittedFlag(ctx context.Context, userID int64, challengeItem *practiceentity.Challenge, flag string) (bool, error) {
	switch challengeItem.FlagType {
	case practiceentity.FlagTypeStatic:
		inputHash := crypto.HashStaticFlag(flag, challengeItem.FlagSalt)
		return crypto.ValidateFlag(inputHash, challengeItem.FlagHash), nil
	case practiceentity.FlagTypeRegex:
		return regexp.MatchString(challengeItem.FlagRegex, flag)
	case practiceentity.FlagTypeManualReview:
		return false, nil
	case practiceentity.FlagTypeDynamic:
	default:
		return false, apperror.ErrInvalidParams.WithCause(fmt.Errorf("unsupported flag type %s", challengeItem.FlagType))
	}

	instance, err := s.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeItem.ID)
	if err != nil {
		return false, apperror.ErrInternal.WithCause(err)
	}
	if instance == nil || instance.Nonce == "" {
		return false, nil
	}
	secret, ok := s.flagSecretForKeyID(instance.FlagKeyID)
	if !ok {
		return false, nil
	}

	expectedFlag := crypto.GenerateDynamicFlag(userID, challengeItem.ID, secret, instance.Nonce, challengeItem.FlagPrefix)
	return crypto.ValidateFlag(flag, expectedFlag), nil
}

func (s *Service) activeFlagSecretKeyID() string {
	if s == nil || s.config == nil {
		return ""
	}
	if keyID := strings.TrimSpace(s.config.Container.ResolvedFlagSecretKeyID); keyID != "" {
		return keyID
	}
	if keyID := strings.TrimSpace(s.config.Container.FlagGlobalSecretKeyID); keyID != "" {
		return keyID
	}
	return "default"
}

func (s *Service) flagSecretForKeyID(keyID string) (string, bool) {
	if s == nil || s.config == nil {
		return "", false
	}
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		keyID = "default"
	}
	if keyID == "" {
		return "", false
	}
	if secret := strings.TrimSpace(s.config.Container.ResolvedFlagSecrets[keyID]); secret != "" {
		return secret, true
	}
	activeKeyID := s.activeFlagSecretKeyID()
	if keyID == activeKeyID || (keyID == "default" && activeKeyID == "") {
		if secret := strings.TrimSpace(s.config.Container.FlagGlobalSecret); secret != "" {
			return secret, true
		}
	}
	return "", false
}
