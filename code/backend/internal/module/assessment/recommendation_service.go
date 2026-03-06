package assessment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type ChallengeRepository interface {
	FindPublishedForRecommendation(limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
}

type RecommendationService struct {
	repo          *Repository
	challengeRepo ChallengeRepository
	redis         *redis.Client
	logger        *zap.Logger
	config        config.RecommendationConfig
	db            *gorm.DB
}

func NewRecommendationService(repo *Repository, challengeRepo ChallengeRepository, redis *redis.Client, cfg config.RecommendationConfig, logger *zap.Logger) *RecommendationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationService{
		repo:          repo,
		challengeRepo: challengeRepo,
		redis:         redis,
		logger:        logger,
		config:        normalizeRecommendationConfig(cfg),
		db:            repo.db,
	}
}

func (s *RecommendationService) GetWeakDimensions(userID int64) ([]string, error) {
	profiles, err := s.repo.FindByUserID(userID)
	if err != nil {
		s.logger.Error("查询能力画像失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	weakDimensions := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Score < s.config.WeakThreshold {
			weakDimensions = append(weakDimensions, profile.Dimension)
		}
	}

	return weakDimensions, nil
}

func (s *RecommendationService) Recommend(userID int64, limit int) (*dto.RecommendationResp, error) {
	weakDimensions, err := s.GetWeakDimensions(userID)
	if err != nil {
		return nil, err
	}

	recommendations, err := s.RecommendChallenges(userID, limit)
	if err != nil {
		return nil, err
	}

	return &dto.RecommendationResp{
		WeakDimensions: weakDimensions,
		Challenges:     recommendations,
	}, nil
}

func (s *RecommendationService) RecommendChallenges(userID int64, limit int) ([]*dto.ChallengeRecommendation, error) {
	if limit <= 0 {
		limit = s.config.DefaultLimit
	}
	if limit > s.config.MaxLimit {
		limit = s.config.MaxLimit
	}

	ctx := context.Background()
	cacheKey := rediskeys.RecommendationKey(userID)
	useCache := limit == s.config.DefaultLimit
	if useCache && s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var recommendations []*dto.ChallengeRecommendation
			if err := json.Unmarshal([]byte(cached), &recommendations); err == nil {
				return recommendations, nil
			}
			s.logger.Warn("推荐缓存反序列化失败", zap.String("cache_key", cacheKey), zap.Error(err))
		}
	}

	weakDimensions, err := s.GetWeakDimensions(userID)
	if err != nil {
		return nil, err
	}
	if len(weakDimensions) == 0 {
		return []*dto.ChallengeRecommendation{}, nil
	}

	solvedIDs, err := s.getSolvedChallengeIDs(userID)
	if err != nil {
		s.logger.Error("查询已解题目失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	challenges, err := s.challengeRepo.FindPublishedForRecommendation(limit, weakDimensions, solvedIDs)
	if err != nil {
		s.logger.Error("查询推荐靶场失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	recommendations := make([]*dto.ChallengeRecommendation, 0, len(challenges))
	for _, challenge := range challenges {
		recommendations = append(recommendations, &dto.ChallengeRecommendation{
			ID:         challenge.ID,
			Title:      challenge.Title,
			Category:   challenge.Category,
			Difficulty: challenge.Difficulty,
			Points:     challenge.Points,
			Reason:     fmt.Sprintf("针对薄弱维度：%s", strings.ToUpper(challenge.Category)),
		})
	}

	if useCache && s.redis != nil {
		if data, err := json.Marshal(recommendations); err == nil {
			if err := s.redis.Set(ctx, cacheKey, data, s.config.CacheTTL).Err(); err != nil {
				s.logger.Warn("推荐缓存写入失败", zap.String("cache_key", cacheKey), zap.Error(err))
			}
		} else {
			s.logger.Error("推荐结果序列化失败", zap.Error(err))
		}
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

func normalizeRecommendationConfig(cfg config.RecommendationConfig) config.RecommendationConfig {
	if cfg.WeakThreshold < 0 || cfg.WeakThreshold > 1 {
		cfg.WeakThreshold = 0.4
	}
	if cfg.CacheTTL < time.Minute {
		cfg.CacheTTL = time.Hour
	}
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 6
	}
	if cfg.MaxLimit < cfg.DefaultLimit {
		cfg.MaxLimit = 20
	}
	return cfg
}
