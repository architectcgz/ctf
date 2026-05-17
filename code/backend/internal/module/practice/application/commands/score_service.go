package commands

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ScoreService struct {
	repo       practiceScoreRepository
	stateStore practiceports.PracticeScoreStateStore
	logger     *zap.Logger
	config     *config.ScoreConfig
}

type practiceScoreRepository interface {
	practiceports.PracticeChallengeScoreRepository
	practiceports.PracticeSolvedChallengeRepository
	practiceports.PracticeUserScoreWriteRepository
}

func NewScoreService(repo practiceScoreRepository, stateStore practiceports.PracticeScoreStateStore, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.ScoreConfig{}
	}
	return &ScoreService{
		repo:       repo,
		stateStore: stateStore,
		logger:     logger,
		config:     cfg,
	}
}

func (s *ScoreService) CalculateScore(ctx context.Context, challengeID int64) int {
	challenge, err := s.repo.FindChallengeScore(ctx, challengeID)
	if err != nil {
		s.logger.Error("查询题目失败", zap.Int64("challengeID", challengeID), zap.Error(err))
		return 0
	}

	return domain.CalculateChallengeScore(challenge)
}

func (s *ScoreService) UpdateUserScore(ctx context.Context, userID int64) error {
	var lock practiceports.PracticeScoreLockLease
	if s.stateStore != nil {
		acquiredLock, acquired, err := s.stateStore.AcquireUserScoreUpdateLock(ctx, userID, s.lockTimeout())
		if err != nil {
			s.logger.Error("获取计分锁失败", zap.Int64("userID", userID), zap.Error(err))
			return fmt.Errorf("获取分布式锁失败: %w", err)
		}
		if !acquired {
			s.logger.Warn("计分锁已被占用", zap.Int64("userID", userID))
			return fmt.Errorf("用户 %d 正在计分中，请稍后重试", userID)
		}
		lock = acquiredLock
	}

	if lock != nil {
		defer func() {
			released, err := lock.Release(ctx)
			if err != nil {
				s.logger.Error("释放分布式锁失败",
					zap.Int64("userID", userID),
					zap.String("lockKey", lock.Key(ctx)),
					zap.Error(err))
			} else if !released {
				s.logger.Warn("锁已被其他协程占用或已过期",
					zap.Int64("userID", userID),
					zap.String("lockKey", lock.Key(ctx)))
			}
		}()
	}

	challengeIDs, err := s.repo.ListSolvedChallengeIDs(ctx, userID)
	if err != nil {
		return err
	}

	challenges, err := s.repo.FindChallengesScores(ctx, challengeIDs)
	if err != nil {
		return err
	}

	totalScore := 0
	for _, challenge := range challenges {
		totalScore += domain.CalculateChallengeScore(&challenge)
	}

	err = s.repo.UpsertUserScore(ctx, &model.UserScore{
		UserID:      userID,
		TotalScore:  totalScore,
		SolvedCount: len(challengeIDs),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return err
	}

	s.logger.Info("更新用户得分", zap.Int64("userID", userID), zap.Int("totalScore", totalScore), zap.Int("solvedCount", len(challengeIDs)))

	info := &dto.UserScoreInfo{
		UserID:      userID,
		TotalScore:  totalScore,
		SolvedCount: len(challengeIDs),
	}
	if s.stateStore != nil {
		if err := s.stateStore.SyncUserScoreState(ctx, info, s.cacheTTL()); err != nil {
			s.logger.Error("批量更新缓存失败", zap.Int64("userID", userID), zap.Error(err))
			return fmt.Errorf("更新得分成功但缓存同步失败: %w", err)
		}
	}

	return nil
}

func (s *ScoreService) lockTimeout() time.Duration {
	if s.config == nil || s.config.LockTimeout <= 0 {
		return 0
	}
	return s.config.LockTimeout
}

func (s *ScoreService) cacheTTL() time.Duration {
	if s.config == nil || s.config.CacheTTL <= 0 {
		return 0
	}
	return s.config.CacheTTL
}
