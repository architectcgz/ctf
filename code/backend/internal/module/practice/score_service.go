package practice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/pkg/cache"
)

// 难度权重映射
var difficultyWeights = map[string]float64{
	model.ChallengeDifficultyBeginner: 1.0,
	model.ChallengeDifficultyEasy:     1.2,
	model.ChallengeDifficultyMedium:   1.5,
	model.ChallengeDifficultyHard:     2.0,
	model.ChallengeDifficultyInsane:   3.0,
}

// ScoreService 计分服务
type ScoreService struct {
	repo   *Repository
	redis  *redis.Client
	logger *zap.Logger
	config *config.ScoreConfig
}

func NewScoreService(repo *Repository, redis *redis.Client, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
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

// CalculateScore 计算题目得分
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

	return calculateChallengeScore(challenge)
}

// UpdateUserScore 更新用户总分
func (s *ScoreService) UpdateUserScore(userID int64) error {
	return s.UpdateUserScoreWithContext(context.Background(), userID)
}

func (s *ScoreService) UpdateUserScoreWithContext(ctx context.Context, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	lockKey := cache.ScoreLockKey(userID)

	// 获取分布式锁（使用唯一 token 防止误删）
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

	// 释放锁时验证 token，避免误删其他协程的锁
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

	// 查询用户已解决的题目
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
		totalScore += calculateChallengeScore(&challenge)
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

	// 使用 Pipeline 批量更新 Redis
	pipe := s.redis.Pipeline()
	cacheKey := cache.UserScoreKey(userID)

	// 缓存完整的 UserScoreInfo JSON（与 GetUserScore 保持一致）
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

// GetUserScore 获取用户得分信息
func (s *ScoreService) GetUserScore(userID int64) (*dto.UserScoreInfo, error) {
	return s.GetUserScoreWithContext(context.Background(), userID)
}

func (s *ScoreService) GetUserScoreWithContext(ctx context.Context, userID int64) (*dto.UserScoreInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cacheKey := cache.UserScoreKey(userID)

	// 尝试从缓存获取
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var info dto.UserScoreInfo
		if json.Unmarshal([]byte(cached), &info) == nil {
			return &info, nil
		}
	}

	// 查询数据库
	userScore, err := s.repo.FindUserScoreWithContext(ctx, userID)

	if err == gorm.ErrRecordNotFound {
		// 空结果不缓存，直接返回默认值
		return &dto.UserScoreInfo{
			UserID:      userID,
			TotalScore:  0,
			SolvedCount: 0,
			Rank:        0,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	userMap, userErr := s.getUsernamesWithContext(ctx, []int64{userID})
	if userErr != nil {
		s.logger.Warn("查询用户名失败", zap.Int64("userID", userID), zap.Error(userErr))
	}

	info := &dto.UserScoreInfo{
		UserID:      userScore.UserID,
		Username:    userMap[userID],
		TotalScore:  userScore.TotalScore,
		SolvedCount: userScore.SolvedCount,
		Rank:        userScore.Rank,
	}

	// 只缓存存在的记录
	data, _ := json.Marshal(info)
	s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL)

	return info, nil
}

// GetRanking 获取排行榜
func (s *ScoreService) GetRanking(limit int) ([]*dto.RankingItem, error) {
	return s.GetRankingWithContext(context.Background(), limit)
}

func (s *ScoreService) GetRankingWithContext(ctx context.Context, limit int) ([]*dto.RankingItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 限制查询上限
	if limit <= 0 || limit > s.config.MaxRankingLimit {
		limit = s.config.MaxRankingLimit
	}

	// 尝试从 Redis 获取
	members, err := s.redis.ZRevRangeWithScores(ctx, cache.RankingKey(), 0, int64(limit-1)).Result()
	if err == nil && len(members) > 0 {
		userIDs := make([]int64, 0, len(members))
		for _, member := range members {
			userIDStr, ok := member.Member.(string)
			if !ok {
				s.logger.Error("排行榜数据类型错误", zap.Any("member", member.Member))
				continue
			}
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				s.logger.Error("解析用户ID失败", zap.String("userIDStr", userIDStr), zap.Error(err))
				continue
			}
			userIDs = append(userIDs, userID)
		}

		// 批量查询用户信息
		userMap, err := s.getUsernamesWithContext(ctx, userIDs)
		if err != nil {
			s.logger.Error("批量查询用户名失败", zap.Error(err))
		}

		result := make([]*dto.RankingItem, 0, len(userIDs))
		for i, userID := range userIDs {
			result = append(result, &dto.RankingItem{
				Rank:       i + 1,
				UserID:     userID,
				Username:   userMap[userID],
				TotalScore: int(members[i].Score),
			})
		}
		return result, nil
	}

	// Redis 失败，从数据库查询
	scores, err := s.repo.ListTopUserScoresWithContext(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 批量查询用户信息
	userIDs := make([]int64, len(scores))
	for i, score := range scores {
		userIDs[i] = score.UserID
	}
	userMap, err := s.getUsernamesWithContext(ctx, userIDs)
	if err != nil {
		s.logger.Error("批量查询用户名失败", zap.Error(err))
	}

	result := make([]*dto.RankingItem, 0, len(scores))
	for i, score := range scores {
		result = append(result, &dto.RankingItem{
			Rank:        i + 1,
			UserID:      score.UserID,
			Username:    userMap[score.UserID],
			TotalScore:  score.TotalScore,
			SolvedCount: score.SolvedCount,
		})
	}

	return result, nil
}

// getUsernames 批量查询用户名
func (s *ScoreService) getUsernames(userIDs []int64) (map[int64]string, error) {
	return s.getUsernamesWithContext(context.Background(), userIDs)
}

func (s *ScoreService) getUsernamesWithContext(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(userIDs) == 0 {
		return make(map[int64]string), nil
	}

	users, err := s.repo.FindUsersByIDsWithContext(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(userIDs))
	for _, user := range users {
		result[user.ID] = user.Username
	}

	// 为不存在的用户填充默认值
	for _, id := range userIDs {
		if _, exists := result[id]; !exists {
			result[id] = fmt.Sprintf("用户%d", id)
			s.logger.Warn("用户不存在", zap.Int64("userID", id))
		}
	}

	return result, nil
}

func calculateChallengeScore(challenge *model.Challenge) int {
	if challenge == nil {
		return 0
	}

	weight := difficultyWeights[challenge.Difficulty]
	if weight == 0 {
		weight = 1.0
	}

	return int(float64(challenge.Points) * weight)
}

func (s *ScoreService) lockTimeout() time.Duration {
	if s.config == nil || s.config.LockTimeout <= 0 {
		return 0
	}
	return s.config.LockTimeout
}
