package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/practice/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/pkg/cache"
)

type ScoreService struct {
	repo   practiceports.PracticeScoreRepository
	redis  *redis.Client
	logger *zap.Logger
	config *config.ScoreConfig
}

func NewScoreService(repo practiceports.PracticeScoreRepository, redis *redis.Client, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ScoreService{
		repo:   repo,
		redis:  redis,
		logger: logger,
		config: cfg,
	}
}

func (s *ScoreService) CalculateScore(challengeID int64) int {
	return s.CalculateScoreWithContext(context.Background(), challengeID)
}

func (s *ScoreService) CalculateScoreWithContext(ctx context.Context, challengeID int64) int {
	if ctx == nil {
		ctx = context.Background()
	}

	challenge, err := s.repo.FindChallengeScoreWithContext(ctx, challengeID)
	if err != nil {
		s.logger.Error("查询题目失败", zap.Int64("challengeID", challengeID), zap.Error(err))
		return 0
	}

	return domain.CalculateChallengeScore(challenge)
}

func (s *ScoreService) UpdateUserScore(userID int64) error {
	return s.UpdateUserScoreWithContext(context.Background(), userID)
}

func (s *ScoreService) UpdateUserScoreWithContext(ctx context.Context, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	lockKey := cache.ScoreLockKey(userID)
	lockToken := uuid.New().String()
	lock, err := s.redis.SetNX(ctx, lockKey, lockToken, s.config.LockTimeout).Result()
	if err != nil {
		s.logger.Error("获取计分锁失败", zap.Int64("userID", userID), zap.Error(err))
		return fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !lock {
		s.logger.Warn("计分锁已被占用", zap.Int64("userID", userID))
		return fmt.Errorf("用户 %d 正在计分中，请稍后重试", userID)
	}

	defer func() {
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		result, err := s.redis.Eval(ctx, script, []string{lockKey}, lockToken).Result()
		if err != nil {
			s.logger.Error("释放分布式锁失败",
				zap.Int64("userID", userID),
				zap.String("lockKey", lockKey),
				zap.Error(err))
		} else if result == int64(0) {
			s.logger.Warn("锁已被其他协程占用或已过期",
				zap.Int64("userID", userID),
				zap.String("lockToken", lockToken))
		}
	}()

	challengeIDs, err := s.repo.ListSolvedChallengeIDsWithContext(ctx, userID)
	if err != nil {
		return err
	}

	challenges, err := s.repo.FindChallengesScoresWithContext(ctx, challengeIDs)
	if err != nil {
		return err
	}

	totalScore := 0
	for _, challenge := range challenges {
		totalScore += domain.CalculateChallengeScore(&challenge)
	}

	err = s.repo.UpsertUserScoreWithContext(ctx, &model.UserScore{
		UserID:      userID,
		TotalScore:  totalScore,
		SolvedCount: len(challengeIDs),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return err
	}

	s.logger.Info("更新用户得分", zap.Int64("userID", userID), zap.Int("totalScore", totalScore), zap.Int("solvedCount", len(challengeIDs)))

	pipe := s.redis.Pipeline()
	cacheKey := cache.UserScoreKey(userID)

	info := &dto.UserScoreInfo{
		UserID:      userID,
		TotalScore:  totalScore,
		SolvedCount: len(challengeIDs),
	}
	data, _ := json.Marshal(info)
	pipe.Set(ctx, cacheKey, data, s.config.CacheTTL)

	pipe.ZAdd(ctx, cache.RankingKey(), redis.Z{
		Score:  float64(totalScore),
		Member: userID,
	})

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("批量更新缓存失败", zap.Int64("userID", userID), zap.Error(err))
		return fmt.Errorf("更新得分成功但缓存同步失败: %w", err)
	}

	return nil
}

func (s *ScoreService) lockTimeout() time.Duration {
	if s.config == nil || s.config.LockTimeout <= 0 {
		return 0
	}
	return s.config.LockTimeout
}
