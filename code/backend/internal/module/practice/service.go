package practice

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo          *Repository
	challengeRepo ChallengeRepository
	instanceRepo  InstanceRepository
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

func NewService(repo *Repository, challengeRepo ChallengeRepository, instanceRepo InstanceRepository, globalSecret string, submitLimit int, submitWindow time.Duration) *Service {
	return &Service{
		repo:          repo,
		challengeRepo: challengeRepo,
		instanceRepo:  instanceRepo,
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
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrChallengeNotPublish
	}

	// 2. 检查是否已完成
	_, err = s.repo.FindCorrectSubmission(userID, challengeID)
	if err == nil {
		return &dto.SubmissionResp{
			IsCorrect:   true,
			Message:     "该题目已完成",
			SubmittedAt: time.Now(),
		}, nil
	}

	// 3. 防暴力破解：检查提交频率
	since := time.Now().Add(-s.submitWindow)
	count, err := s.repo.CountRecentSubmissions(userID, challengeID, since)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if count >= int64(s.submitLimit) {
		return nil, errcode.ErrSubmitTooFrequent
	}

	// 4. 验证 Flag
	isCorrect := false
	if challenge.FlagType == model.FlagTypeStatic {
		// 静态 Flag：对比哈希
		inputHash := crypto.HashStaticFlag(flag, challenge.FlagSalt)
		isCorrect = crypto.ValidateFlag(inputHash, challenge.FlagHash)
	} else {
		// 动态 Flag：查找实例并生成 Flag 对比
		instance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
		if err == nil && instance.Nonce != "" {
			expectedFlag := crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, instance.Nonce)
			isCorrect = crypto.ValidateFlag(flag, expectedFlag)
		}
	}

	// 5. 记录提交
	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		Flag:        flag,
		IsCorrect:   isCorrect,
		SubmittedAt: time.Now(),
	}
	if err := s.repo.CreateSubmission(submission); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// 6. 返回结果
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
