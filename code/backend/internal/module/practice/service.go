package practice

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repo          *Repository
	challengeRepo ChallengeRepository
	instanceRepo  InstanceRepository
	redis         *redis.Client
	logger        *zap.Logger
	globalSecret  string
	submitLimit   int
	submitWindow  time.Duration
}

type ChallengeRepository interface {
	FindByID(id int64) (*model.Challenge, error)
}

type InstanceRepository interface {
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

func NewService(repo *Repository, challengeRepo ChallengeRepository, instanceRepo InstanceRepository, redis *redis.Client, logger *zap.Logger, globalSecret string, submitLimit int, submitWindow time.Duration) *Service {
	return &Service{
		repo:          repo,
		challengeRepo: challengeRepo,
		instanceRepo:  instanceRepo,
		redis:         redis,
		logger:        logger,
		globalSecret:  globalSecret,
		submitLimit:   submitLimit,
		submitWindow:  submitWindow,
	}
}

// SubmitFlag 提交 Flag
func (s *Service) SubmitFlag(userID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	// 1. 检查靶场是否存在且已发布
	challenge, err := s.challengeRepo.FindByID(challengeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrChallengeNotFound
		}
		s.logger.Error("查询靶场失败", zap.Int64("challengeID", challengeID), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	// 2. 检查是否已完成
	_, err = s.repo.FindCorrectSubmission(userID, challengeID)
	if err == nil {
		return nil, errcode.ErrAlreadySolved
	}

	// 3. 防暴力破解：使用 Redis 限流（每用户每题）
	rateLimitKey := fmt.Sprintf("ctf:submit:%d:%d", userID, challengeID)
	ctx := context.Background()
	count, err := s.redis.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		s.logger.Error("Redis限流失败", zap.String("key", rateLimitKey), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if count == 1 {
		s.redis.Expire(ctx, rateLimitKey, s.submitWindow)
	}
	if count > int64(s.submitLimit) {
		s.logger.Warn("提交频率超限", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID), zap.Int64("count", count), zap.Int("limit", s.submitLimit))
		return nil, errcode.ErrSubmitTooFrequent
	}

	// 4. 验证 Flag
	isCorrect := false
	if challenge.FlagType == model.FlagTypeStatic {
		inputHash := crypto.HashStaticFlag(flag, challenge.FlagSalt)
		isCorrect = crypto.ValidateFlag(inputHash, challenge.FlagHash)
	} else {
		instance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				isCorrect = false
			} else {
				s.logger.Error("查询实例失败", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID), zap.Error(err))
				return nil, errcode.ErrInternal.WithCause(err)
			}
		} else {
			// 实例存在，验证 Flag
			if instance.Nonce != "" {
				expectedFlag := crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, instance.Nonce)
				isCorrect = crypto.ValidateFlag(flag, expectedFlag)
			}
		}
	}

	// 5. 记录提交（不存储明文 Flag）
	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		Flag:        "",
		IsCorrect:   isCorrect,
		SubmittedAt: time.Now(),
	}
	if err := s.repo.CreateSubmission(submission); err != nil {
		// 处理唯一约束冲突（并发提交导致）
		if s.repo.IsUniqueViolation(err) && isCorrect {
			s.logger.Warn("并发提交检测到重复", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID))
			return nil, errcode.ErrAlreadySolved
		}
		s.logger.Error("创建提交记录失败", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID), zap.Error(err))
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// 6. 记录日志
	if isCorrect {
		s.logger.Info("Flag验证成功", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID))
	} else {
		s.logger.Debug("Flag验证失败", zap.Int64("userID", userID), zap.Int64("challengeID", challengeID), zap.String("flagPrefix", flag[:min(len(flag), 10)]))
	}

	// 7. 返回结果
	resp := &dto.SubmissionResp{
		IsCorrect:   isCorrect,
		SubmittedAt: submission.SubmittedAt,
	}
	if isCorrect {
		resp.Message = "恭喜你，Flag 正确！"
		resp.Points = challenge.Points
	} else {
		resp.Message = "Flag 错误，请重试"
	}

	return resp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
