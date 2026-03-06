package practice

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/pkg/cache"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
	config *config.ScoreConfig
}

func NewScoreService(db *gorm.DB, redis *redis.Client, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
	return &ScoreService{
		db:     db,
		redis:  redis,
		logger: logger,
		config: cfg,
	}
}

// CalculateScore 计算题目得分
func (s *ScoreService) CalculateScore(challengeID int64) int {
	var challenge model.Challenge
	if err := s.db.Select("points, difficulty").Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		s.logger.Error("查询题目失败", zap.Int64("challengeID", challengeID), zap.Error(err))
		return 0
	}

	weight := difficultyWeights[challenge.Difficulty]
	if weight == 0 {
		weight = 1.0
	}

	return int(float64(challenge.Points) * weight)
}

// UpdateUserScore 更新用户总分
func (s *ScoreService) UpdateUserScore(userID int64) error {
	ctx := context.Background()
	lockKey := cache.ScoreLockKey(userID)

	// 获取分布式锁
	lock, err := s.redis.SetNX(ctx, lockKey, 1, s.config.LockTimeout).Result()
	if err != nil || !lock {
		s.logger.Warn("获取计分锁失败", zap.Int64("userID", userID), zap.Error(err))
		return nil
	}
	defer s.redis.Del(ctx, lockKey)

	// 查询用户已解决的题目
	var submissions []model.Submission
	err = s.db.Where("user_id = ? AND is_correct = ?", userID, true).
		Select("DISTINCT challenge_id").
		Find(&submissions).Error
	if err != nil {
		return err
	}

	// 计算总分
	totalScore := 0
	for _, sub := range submissions {
		totalScore += s.CalculateScore(sub.ChallengeID)
	}

	// 更新数据库（使用 GORM Clauses 实现跨数据库兼容）
	err = s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_score", "solved_count", "updated_at"}),
	}).Create(&model.UserScore{
		UserID:      userID,
		TotalScore:  totalScore,
		SolvedCount: len(submissions),
		UpdatedAt:   time.Now(),
	}).Error
	if err != nil {
		return err
	}

	s.logger.Info("更新用户得分", zap.Int64("userID", userID), zap.Int("totalScore", totalScore), zap.Int("solvedCount", len(submissions)))

	// 使用 Pipeline 批量更新 Redis
	pipe := s.redis.Pipeline()
	cacheKey := cache.UserScoreKey(userID)
	pipe.Set(ctx, cacheKey, totalScore, s.config.CacheTTL)
	pipe.ZAdd(ctx, cache.RankingKey(), redis.Z{
		Score:  float64(totalScore),
		Member: userID,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("批量更新缓存失败", zap.Int64("userID", userID), zap.Error(err))
	}

	return nil
}

// GetUserScore 获取用户得分信息
func (s *ScoreService) GetUserScore(userID int64) (*dto.UserScoreInfo, error) {
	ctx := context.Background()
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
	var userScore model.UserScore
	err = s.db.Where("user_id = ?", userID).First(&userScore).Error

	var info *dto.UserScoreInfo
	if err == gorm.ErrRecordNotFound {
		info = &dto.UserScoreInfo{
			UserID:      userID,
			TotalScore:  0,
			SolvedCount: 0,
			Rank:        0,
		}
	} else if err != nil {
		return nil, err
	} else {
		var user model.User
		s.db.Select("username").Where("id = ?", userID).First(&user)

		info = &dto.UserScoreInfo{
			UserID:      userScore.UserID,
			Username:    user.Username,
			TotalScore:  userScore.TotalScore,
			SolvedCount: userScore.SolvedCount,
			Rank:        userScore.Rank,
		}
	}

	// 缓存结果（包括空结果）
	data, _ := json.Marshal(info)
	s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL)

	return info, nil
}

// GetRanking 获取排行榜
func (s *ScoreService) GetRanking(limit int) ([]*dto.RankingItem, error) {
	ctx := context.Background()

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
		userMap, err := s.getUsernames(userIDs)
		if err != nil {
			s.logger.Error("批量查询用户名失败", zap.Error(err))
		}

		result := make([]*dto.RankingItem, 0, len(members))
		for i, member := range members {
			userIDStr, _ := member.Member.(string)
			userID, _ := strconv.ParseInt(userIDStr, 10, 64)

			result = append(result, &dto.RankingItem{
				Rank:       i + 1,
				UserID:     userID,
				Username:   userMap[userID],
				TotalScore: int(member.Score),
			})
		}
		return result, nil
	}

	// Redis 失败，从数据库查询
	var scores []model.UserScore
	err = s.db.Order("total_score DESC, updated_at ASC").Limit(limit).Find(&scores).Error
	if err != nil {
		return nil, err
	}

	// 批量查询用户信息
	userIDs := make([]int64, len(scores))
	for i, score := range scores {
		userIDs[i] = score.UserID
	}
	userMap, err := s.getUsernames(userIDs)
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
	if len(userIDs) == 0 {
		return make(map[int64]string), nil
	}

	var users []model.User
	err := s.db.Select("id, username").Where("id IN ?", userIDs).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(users))
	for _, user := range users {
		result[user.ID] = user.Username
	}
	return result, nil
}
