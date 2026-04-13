package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	rediskeys "ctf-platform/internal/pkg/redis"
	platformevents "ctf-platform/internal/platform/events"
)

type RecommendationService struct {
	repo          assessmentports.RecommendationRepository
	challengeRepo assessmentports.ChallengeRepository
	redis         *redis.Client
	logger        *zap.Logger
	config        config.RecommendationConfig
}

var _ assessmentcontracts.RecommendationProvider = (*RecommendationService)(nil)

func NewRecommendationService(repo assessmentports.RecommendationRepository, challengeRepo assessmentports.ChallengeRepository, redis *redis.Client, cfg config.RecommendationConfig, logger *zap.Logger) *RecommendationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RecommendationService{
		repo:          repo,
		challengeRepo: challengeRepo,
		redis:         redis,
		logger:        logger,
		config:        assessmentdomain.NormalizeRecommendationConfig(cfg),
	}
}

func (s *RecommendationService) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handlePracticeCacheRefreshEvent)
}

func (s *RecommendationService) RegisterContestEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(contestcontracts.EventAWDAttackAccepted, s.handleContestCacheRefreshEvent)
}

func (s *RecommendationService) handlePracticeCacheRefreshEvent(ctx context.Context, evt platformevents.Event) error {
	if s.redis == nil {
		return nil
	}

	var userID int64
	switch payload := evt.Payload.(type) {
	case practicecontracts.FlagAcceptedEvent:
		userID = payload.UserID
	default:
		return fmt.Errorf("unexpected practice cache refresh payload: %T", evt.Payload)
	}
	if userID <= 0 {
		return nil
	}
	return s.redis.Del(ctx, rediskeys.RecommendationKey(userID)).Err()
}

func (s *RecommendationService) handleContestCacheRefreshEvent(ctx context.Context, evt platformevents.Event) error {
	if s.redis == nil {
		return nil
	}

	payload, ok := evt.Payload.(contestcontracts.AWDAttackAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest awd cache refresh payload: %T", evt.Payload)
	}
	if payload.UserID <= 0 {
		return nil
	}
	return s.redis.Del(ctx, rediskeys.RecommendationKey(payload.UserID)).Err()
}

func (s *RecommendationService) GetWeakDimensions(userID int64) ([]string, error) {
	return s.GetWeakDimensionsWithContext(context.Background(), userID)
}

func (s *RecommendationService) GetWeakDimensionsWithContext(ctx context.Context, userID int64) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	profiles, err := s.repo.FindByUserIDWithContext(ctx, userID)
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
	return s.RecommendWithContext(context.Background(), userID, limit)
}

func (s *RecommendationService) RecommendWithContext(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error) {
	weakDimensions, err := s.GetWeakDimensionsWithContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	recommendations, err := s.RecommendChallengesWithContext(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	return &dto.RecommendationResp{
		WeakDimensions: weakDimensions,
		Challenges:     recommendations,
	}, nil
}

func (s *RecommendationService) RecommendChallenges(userID int64, limit int) ([]*dto.ChallengeRecommendation, error) {
	return s.RecommendChallengesWithContext(context.Background(), userID, limit)
}

func (s *RecommendationService) RecommendChallengesWithContext(ctx context.Context, userID int64, limit int) ([]*dto.ChallengeRecommendation, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if limit <= 0 {
		limit = s.config.DefaultLimit
	}
	if limit > s.config.MaxLimit {
		limit = s.config.MaxLimit
	}

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

	weakDimensions, err := s.GetWeakDimensionsWithContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(weakDimensions) == 0 {
		return []*dto.ChallengeRecommendation{}, nil
	}

	solvedIDs, err := s.getSolvedChallengeIDs(ctx, userID)
	if err != nil {
		s.logger.Error("查询已解题目失败", zap.Int64("user_id", userID), zap.Error(err))
		return nil, err
	}

	challenges, err := s.challengeRepo.FindPublishedForRecommendationWithContext(ctx, limit, weakDimensions, solvedIDs)
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

func (s *RecommendationService) getSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.repo.ListSolvedChallengeIDsWithContext(ctx, userID)
}

type RecommendationQuery struct {
	Limit int `form:"limit"`
}
