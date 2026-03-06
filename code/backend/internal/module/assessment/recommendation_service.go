package assessment

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RecommendationService struct {
	repo             *Repository
	challengeRepo    ChallengeRepository
	redis            *redis.Client
	logger           *zap.Logger
	weakThreshold    float64
	cacheTTL         time.Duration
	cacheKeyPrefix   string
	db               *gorm.DB
}

type ChallengeRepository interface {
	FindPublishedWithTags(limit int, tagIDs []int64, excludeSolved []int64) ([]*model.Challenge, error)
	FindTagsByDimensions(dimensions []string) ([]int64, error)
}

func NewRecommendationService(repo *Repository, challengeRepo ChallengeRepository, redis *redis.Client, logger *zap.Logger, weakThreshold float64, cacheTTL time.Duration, cacheKeyPrefix string) *RecommendationService {
	return &RecommendationService{
		repo:           repo,
		challengeRepo:  challengeRepo,
		redis:          redis,
		logger:         logger,
		weakThreshold:  weakThreshold,
		cacheTTL:       cacheTTL,
		cacheKeyPrefix: cacheKeyPrefix,
		db:             repo.db,
	}
}

func (s *RecommendationService) GetWeakDimensions(userID int64) ([]string, error) {
	profiles, err := s.repo.FindByUserID(userID)
	if err != nil {
		s.logger.Error("查询能力画像失败", zap.Int64("userID", userID), zap.Error(err))
		return nil, err
	}

	var weakDimensions []string
	for _, p := range profiles {
		if p.Score < s.weakThreshold {
			weakDimensions = append(weakDimensions, p.Dimension)
		}
	}

	return weakDimensions, nil
}

// toChallengeRecommendation 将 Model 转换为 DTO
func toChallengeRecommendation(c *model.Challenge, reason string) *dto.ChallengeRecommendation {
	return &dto.ChallengeRecommendation{
		ID:         c.ID,
		Title:      c.Title,
		Category:   c.Category,
		Difficulty: c.Difficulty,
		Points:     c.Points,
		Reason:     reason,
	}
}

func (s *RecommendationService) RecommendChallenges(userID int64, limit int) ([]*dto.ChallengeRecommendation, error) {
	ctx := context.Background()
	cacheKey := rediskeys.RecommendationKey(userID)

	// 尝试从缓存获取
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var recommendations []*dto.ChallengeRecommendation
		if err := json.Unmarshal([]byte(cached), &recommendations); err == nil {
			return recommendations, nil
		} else {
			s.logger.Warn("缓存反序列化失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	// 获取薄弱维度
	weakDimensions, err := s.GetWeakDimensions(userID)
	if err != nil {
		return nil, err
	}

	if len(weakDimensions) == 0 {
		return []*dto.ChallengeRecommendation{}, nil
	}

	// 查询薄弱维度对应的标签 ID
	tagIDs, err := s.challengeRepo.FindTagsByDimensions(weakDimensions)
	if err != nil {
		s.logger.Error("查询标签失败", zap.Strings("dimensions", weakDimensions), zap.Error(err))
		return nil, err
	}

	if len(tagIDs) == 0 {
		return []*dto.ChallengeRecommendation{}, nil
	}

	// 查询用户已解决的题目
	solvedIDs, err := s.getSolvedChallengeIDs(userID)
	if err != nil {
		s.logger.Error("查询已解决题目失败", zap.Int64("userID", userID), zap.Error(err))
		return nil, err
	}

	// 推荐匹配薄弱维度的靶场
	challenges, err := s.challengeRepo.FindPublishedWithTags(limit, tagIDs, solvedIDs)
	if err != nil {
		s.logger.Error("查询推荐靶场失败", zap.Error(err))
		return nil, err
	}

	recommendations := make([]*dto.ChallengeRecommendation, 0, len(challenges))
	for _, c := range challenges {
		reason := fmt.Sprintf("针对薄弱维度：%s", c.Category)
		recommendations = append(recommendations, toChallengeRecommendation(c, reason))
	}

	// 缓存结果
	if data, err := json.Marshal(recommendations); err == nil {
		if err := s.redis.Set(ctx, cacheKey, data, s.cacheTTL).Err(); err != nil {
			s.logger.Warn("缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	} else {
		s.logger.Error("推荐结果序列化失败", zap.Error(err))
	}

	return recommendations, nil
}

func (s *RecommendationService) getSolvedChallengeIDs(userID int64) ([]int64, error) {
	var ids []int64
	err := s.db.Model(&model.Submission{}).
		Where("user_id = ? AND is_correct = ?", userID, true).
		Distinct("challenge_id").
		Pluck("challenge_id", &ids).Error
	return ids, err
}
