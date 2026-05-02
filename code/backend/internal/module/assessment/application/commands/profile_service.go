package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/pkg/errcode"
)

type Service struct {
	repo   assessmentports.ProfileRepository
	redis  *redis.Client
	config config.AssessmentConfig
	logger *zap.Logger
}

var _ assessmentcontracts.ProfileService = (*Service)(nil)

func NewProfileService(repo assessmentports.ProfileRepository, redis *redis.Client, cfg config.AssessmentConfig, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repo:   repo,
		redis:  redis,
		config: assessmentdomain.NormalizeAssessmentConfig(cfg),
		logger: logger,
	}
}

func (s *Service) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handleFlagAcceptedEvent)
}

func (s *Service) RegisterContestEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(contestcontracts.EventAWDAttackAccepted, s.handleAWDAttackAcceptedEvent)
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

func (s *Service) handleAWDAttackAcceptedEvent(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AWDAttackAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest awd event payload: %T", evt.Payload)
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
	score, err := s.repo.GetDimensionScore(ctx, userID, dimension)
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

	return s.repo.Upsert(ctx, profile)
}

// CalculateSkillProfile 带超时控制的画像计算
func (s *Service) CalculateSkillProfile(ctx context.Context, userID int64) ([]*dto.SkillDimension, error) {
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
		return s.getExistingProfile(ctx, userID)
	}
	if !locked {
		s.logger.Debug("画像正在计算中，跳过", zap.Int64("userID", userID))
		return s.getExistingProfile(ctx, userID)
	}
	defer s.unlock(ctx, lockKey)

	// 调用 Repository 查询维度得分
	scores, err := s.repo.GetDimensionScores(ctx, userID)
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
	if err := s.repo.BatchUpsert(ctx, profiles); err != nil {
		return nil, err
	}

	s.logger.Debug("画像计算结果", zap.Int64("userID", userID), zap.Int("dimensionCount", len(dimensions)))
	return dimensions, nil
}

func (s *Service) RebuildAllSkillProfiles(ctx context.Context) error {
	userIDs, err := s.repo.ListStudentIDs(ctx)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := s.CalculateSkillProfile(ctx, userID); err != nil {
			s.logger.Error("重建能力画像失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	return nil
}

// getExistingProfile 获取已有画像
func (s *Service) getExistingProfile(ctx context.Context, userID int64) ([]*dto.SkillDimension, error) {
	profiles, err := s.repo.FindByUserID(ctx, userID)
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
func (s *Service) GetSkillProfile(ctx context.Context, userID int64) (*dto.SkillProfileResp, error) {
	profiles, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return assessmentdomain.BuildEmptyProfile(userID), nil
	}

	return assessmentdomain.BuildSkillProfile(userID, profiles), nil
}

func (s *Service) GetStudentSkillProfile(ctx context.Context, requesterID int64, requesterRole string, studentID int64) (*dto.SkillProfileResp, error) {
	student, err := s.repo.FindUserByID(ctx, studentID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if student == nil || student.Role != model.RoleStudent {
		return nil, errcode.ErrNotFound
	}

	if requesterRole != model.RoleAdmin {
		requester, findErr := s.repo.FindUserByID(ctx, requesterID)
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

	return s.GetSkillProfile(ctx, studentID)
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
