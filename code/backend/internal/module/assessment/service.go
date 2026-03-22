package assessment

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo   *Repository
	redis  *redis.Client
	config config.AssessmentConfig
	logger *zap.Logger
}

func NewService(repo *Repository, redis *redis.Client, cfg config.AssessmentConfig, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repo:   repo,
		redis:  redis,
		config: normalizeConfig(cfg),
		logger: logger,
	}
}

func (s *Service) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handleFlagAcceptedEvent)
}

func (s *Service) handleFlagAcceptedEvent(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", evt.Payload)
	}
	if !model.IsValidDimension(payload.Dimension) {
		return nil
	}

	updateCtx := ctx
	cancel := func() {}
	if timeout := s.config.IncrementalUpdateTimeout; timeout > 0 {
		updateCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	return s.UpdateSkillProfileForDimension(updateCtx, payload.UserID, payload.Dimension)
}

// CalculateSkillProfile 计算用户能力画像
func (s *Service) CalculateSkillProfile(userID int64) ([]*dto.SkillDimension, error) {
	return s.CalculateSkillProfileWithContext(context.Background(), userID)
}

// UpdateSkillProfileForDimension 增量更新指定维度的能力画像
func (s *Service) UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error {
	// 校验维度合法性
	if !model.IsValidDimension(dimension) {
		s.logger.Warn("无效维度", zap.String("dimension", dimension))
		return fmt.Errorf("invalid dimension: %s", dimension)
	}

	// 使用分布式锁避免并发重复计算
	lockKey := fmt.Sprintf("%s:lock:%d:%s", s.config.RedisKeyPrefix, userID, dimension)
	locked, err := s.tryLock(ctx, lockKey)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		s.logger.Warn("获取分布式锁失败", zap.Int64("userID", userID), zap.String("dimension", dimension), zap.Error(err))
		return err
	}
	if !locked {
		s.logger.Debug("维度画像正在计算中，跳过", zap.Int64("userID", userID), zap.String("dimension", dimension))
		return nil
	}
	defer s.unlock(ctx, lockKey)

	// 查询该维度得分
	score, err := s.repo.GetDimensionScoreWithContext(ctx, userID, dimension)
	if err != nil {
		return err
	}

	var rate float64
	if score.TotalScore > 0 {
		rate = float64(score.UserScore) / float64(score.TotalScore)
	}

	// 更新数据库
	profile := &model.SkillProfile{
		UserID:    userID,
		Dimension: dimension,
		Score:     rate,
		UpdatedAt: time.Now(),
	}

	return s.repo.UpsertWithContext(ctx, profile)
}

// CalculateSkillProfileWithContext 带超时控制的画像计算
func (s *Service) CalculateSkillProfileWithContext(ctx context.Context, userID int64) ([]*dto.SkillDimension, error) {
	start := time.Now()
	defer func() {
		s.logger.Info("画像计算完成", zap.Int64("userID", userID), zap.Duration("duration", time.Since(start)))
	}()

	// 使用分布式锁避免并发重复计算
	lockKey := fmt.Sprintf("%s:lock:%d", s.config.RedisKeyPrefix, userID)
	locked, err := s.tryLock(ctx, lockKey)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		// Redis 故障时返回已有画像，避免进入临界区
		s.logger.Warn("获取分布式锁失败，返回已有画像", zap.Int64("userID", userID), zap.Error(err))
		return s.getExistingProfileWithContext(ctx, userID)
	}
	if !locked {
		s.logger.Debug("画像正在计算中，跳过", zap.Int64("userID", userID))
		return s.getExistingProfileWithContext(ctx, userID)
	}
	defer s.unlock(ctx, lockKey)

	// 调用 Repository 查询维度得分
	scores, err := s.repo.GetDimensionScoresWithContext(ctx, userID)
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
	if err := s.repo.BatchUpsertWithContext(ctx, profiles); err != nil {
		return nil, err
	}

	s.logger.Debug("画像计算结果", zap.Int64("userID", userID), zap.Int("dimensionCount", len(dimensions)))
	return dimensions, nil
}

func (s *Service) RebuildAllSkillProfiles(ctx context.Context) error {
	userIDs, err := s.repo.ListStudentIDsWithContext(ctx)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := s.CalculateSkillProfileWithContext(ctx, userID); err != nil {
			s.logger.Error("重建能力画像失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	return nil
}

// getExistingProfile 获取已有画像
func (s *Service) getExistingProfile(userID int64) ([]*dto.SkillDimension, error) {
	return s.getExistingProfileWithContext(context.Background(), userID)
}

func (s *Service) getExistingProfileWithContext(ctx context.Context, userID int64) ([]*dto.SkillDimension, error) {
	profiles, err := s.repo.FindByUserIDWithContext(ctx, userID)
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
	return s.GetSkillProfileWithContext(context.Background(), userID)
}

func (s *Service) GetSkillProfileWithContext(ctx context.Context, userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserIDWithContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return buildEmptyProfile(userID), nil
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

func (s *Service) GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.SkillProfileResp, error) {
	student, err := s.repo.FindUserByIDWithContext(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, findErr := s.repo.FindUserByIDWithContext(ctx, requesterID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		if requester == nil {
			return nil, errcode.ErrUnauthorized
		}
		if strings.TrimSpace(requester.ClassName) == "" || requester.ClassName != student.ClassName {
			return nil, errcode.ErrForbidden
		}
	}

	return s.GetSkillProfileWithContext(ctx, studentID)
}

func (s *Service) tryLock(ctx context.Context, key string) (bool, error) {
	if s.redis == nil {
		return true, nil
	}
	return s.redis.SetNX(ctx, key, 1, s.config.LockTTL).Result()
}

func (s *Service) unlock(ctx context.Context, key string) {
	if s.redis == nil {
		return
	}
	if err := s.redis.Del(ctx, key).Err(); err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Warn("释放画像锁失败", zap.String("key", key), zap.Error(err))
	}
}

func buildEmptyProfile(userID int64) *dto.SkillProfileResp {
	dimensions := make([]*dto.SkillDimension, 0, len(model.AllDimensions))
	for _, dim := range model.AllDimensions {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: dim,
			Score:     0,
		})
	}

	return &dto.SkillProfileResp{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  "",
	}
}

func normalizeConfig(cfg config.AssessmentConfig) config.AssessmentConfig {
	if cfg.RedisKeyPrefix == "" {
		cfg.RedisKeyPrefix = "ctf:assessment:skill-profile"
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 10 * time.Second
	}
	if cfg.FullRebuildCron == "" {
		cfg.FullRebuildCron = "0 0 * * *"
	}
	if cfg.FullRebuildTimeout <= 0 {
		cfg.FullRebuildTimeout = 30 * time.Minute
	}
	if cfg.IncrementalUpdateDelay <= 0 {
		cfg.IncrementalUpdateDelay = 100 * time.Millisecond
	}
	if cfg.IncrementalUpdateTimeout <= 0 {
		cfg.IncrementalUpdateTimeout = 5 * time.Second
	}
	return cfg
}
