package assessment

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repo   *Repository
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

func NewService(repo *Repository, db *gorm.DB, redis *redis.Client, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CalculateSkillProfile 计算用户能力画像
func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
	return s.CalculateSkillProfileWithContext(context.Background(), userID)
}

// CalculateSkillProfileWithContext 带超时控制的画像计算
func (s *Service) CalculateSkillProfileWithContext(ctx context.Context, userID int64) ([]*dto.SkillDimension, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("画像计算完成", zap.Int64("userID", userID), zap.Duration("duration", time.Since(start)))
	}()

	// 使用分布式锁避免并发重复计算
	lockKey := fmt.Sprintf("skill_profile:lock:%d", userID)
	locked, err := s.redis.SetNX(ctx, lockKey, 1, 10*time.Second).Result()
	if err != nil {
		s.logger.Warn("获取分布式锁失败", zap.Int64("userID", userID), zap.Error(err))
	} else if !locked {
		s.logger.Debug("画像正在计算中，跳过", zap.Int64("userID", userID))
		return s.getExistingProfile(userID)
	}
	defer s.redis.Del(ctx, lockKey)

	// 调用 Repository 查询维度得分
	scores, err := s.repo.GetDimensionScores(userID)
	if err != nil {
		return nil, err
	}

	dimensions := make([]*dto.SkillDimension, 0, len(scores))
	profiles := make([]*model.SkillProfile, 0, len(scores))
	now := time.Now()

	for _, score := range scores {
		// 校验维度合法性
		if !model.IsValidDimension(score.Dimension) {
			s.logger.Warn("跳过无效维度", zap.String("dimension", score.Dimension))
			continue
		}

		var rate float64
		if score.TotalScore > 0 {
			rate = float64(score.UserScore) / float64(score.TotalScore)
		}

		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: score.Dimension,
			Score:     rate,
		})

		profiles = append(profiles, &model.SkillProfile{
			UserID:    userID,
			Dimension: score.Dimension,
			Score:     rate,
			UpdatedAt: now,
		})
	}

	// 保存到数据库
	if err := s.repo.BatchUpsert(profiles); err != nil {
		return nil, err
	}

	s.logger.Debug("画像计算结果", zap.Int64("userID", userID), zap.Int("dimensionCount", len(dimensions)))
	return dimensions, nil
}

// getExistingProfile 获取已有画像
func (s *Service) getExistingProfile(userID int64) ([]*dto.SkillDimension, error) {
	profiles, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	dimensions := make([]*dto.SkillDimension, 0, len(profiles))
	for _, p := range profiles {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: p.Dimension,
			Score:     p.Score,
		})
	}
	return dimensions, nil
}

// GetSkillProfile 获取用户能力画像
func (s *Service) GetSkillProfile(userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 构建维度得分映射
	dimensionMap := make(map[string]float64)
	var latestUpdate time.Time

	for _, p := range profiles {
		dimensionMap[p.Dimension] = p.Score
		if p.UpdatedAt.After(latestUpdate) {
			latestUpdate = p.UpdatedAt
		}
	}

	// 填充所有维度（缺失的默认为 0）
	dimensions := make([]*dto.SkillDimension, 0, len(model.AllDimensions))
	for _, dim := range model.AllDimensions {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: dim,
			Score:     dimensionMap[dim], // 不存在时为 0
		})
	}

	return &dto.SkillProfileResp{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  latestUpdate.Format(time.RFC3339),
	}, nil
}
