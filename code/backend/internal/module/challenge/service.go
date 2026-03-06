package challenge

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/cache"
	"ctf-platform/pkg/errcode"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	imageRepo *ImageRepository
	redis     *redis.Client
	config    *Config
	log       *zap.Logger
}

func NewService(repo *Repository, imageRepo *ImageRepository, redis *redis.Client, config *Config, log *zap.Logger) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	return &Service{
		repo:      repo,
		imageRepo: imageRepo,
		redis:     redis,
		config:    config,
		log:       log,
	}
}

func (s *Service) CreateChallenge(req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	// 验证镜像是否存在
	_, err := s.imageRepo.FindByID(req.ImageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound.WithCause(errors.New(ErrMsgImageNotFound))
		}
		return nil, err
	}

	challenge := &model.Challenge{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		Points:      req.Points,
		ImageID:     req.ImageID,
		Status:      model.ChallengeStatusDraft,
	}

	if err := s.repo.Create(challenge); err != nil {
		return nil, err
	}

	return s.toResp(challenge), nil
}

func (s *Service) UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if req.Title != "" {
		challenge.Title = req.Title
	}
	if req.Description != "" {
		challenge.Description = req.Description
	}
	if req.Category != "" {
		challenge.Category = req.Category
	}
	if req.Difficulty != "" {
		challenge.Difficulty = req.Difficulty
	}
	if req.Points > 0 {
		challenge.Points = req.Points
	}
	if req.ImageID > 0 {
		_, err := s.imageRepo.FindByID(req.ImageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrNotFound.WithCause(errors.New(ErrMsgImageNotFound))
			}
			return err
		}
		challenge.ImageID = req.ImageID
	}

	return s.repo.Update(challenge)
}

func (s *Service) DeleteChallenge(id int64) error {
	hasInstances, err := s.repo.HasRunningInstances(id)
	if err != nil {
		return err
	}
	if hasInstances {
		return errcode.ErrConflict.WithCause(errors.New(ErrMsgHasRunningInstances))
	}

	return s.repo.Delete(id)
}

func (s *Service) GetChallenge(id int64) (*dto.ChallengeResp, error) {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toResp(challenge), nil
}

func (s *Service) ListChallenges(query *dto.ChallengeQuery) (*dto.PageResult, error) {
	challenges, total, err := s.repo.List(query)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ChallengeResp, len(challenges))
	for i, c := range challenges {
		list[i] = s.toResp(c)
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = 20
	}

	return &dto.PageResult{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *Service) PublishChallenge(id int64) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if challenge.ImageID == 0 {
		return errcode.ErrInvalidParams.WithCause(errors.New(ErrMsgImageNotConfigured))
	}

	challenge.Status = model.ChallengeStatusPublished
	return s.repo.Update(challenge)
}

func (s *Service) toResp(c *model.Challenge) *dto.ChallengeResp {
	return &dto.ChallengeResp{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Category:    c.Category,
		Difficulty:  c.Difficulty,
		Points:      c.Points,
		ImageID:     c.ImageID,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ListPublishedChallenges 查询已发布靶场列表（学员视图）
func (s *Service) ListPublishedChallenges(userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	challenges, total, err := s.repo.ListPublished(query)
	if err != nil {
		return nil, err
	}

	if len(challenges) == 0 {
		return &dto.PageResult{
			List:  []*dto.ChallengeListItem{},
			Total: total,
			Page:  query.Page,
			Size:  query.Size,
		}, nil
	}

	// 批量查询，解决 N+1 问题
	challengeIDs := make([]int64, len(challenges))
	for i, c := range challenges {
		challengeIDs[i] = c.ID
	}

	solvedMap := make(map[int64]bool)
	if userID > 0 {
		solvedMap, err = s.repo.BatchGetSolvedStatus(userID, challengeIDs)
		if err != nil {
			s.log.Error("failed to batch get solved status", zap.Error(err))
		}
	}

	solvedCountMap, err := s.repo.BatchGetSolvedCount(challengeIDs)
	if err != nil {
		s.log.Error("failed to batch get solved count", zap.Error(err))
	}

	attemptsMap, err := s.repo.BatchGetTotalAttempts(challengeIDs)
	if err != nil {
		s.log.Error("failed to batch get total attempts", zap.Error(err))
	}

	list := make([]*dto.ChallengeListItem, 0, len(challenges))
	for _, c := range challenges {
		list = append(list, &dto.ChallengeListItem{
			ID:            c.ID,
			Title:         c.Title,
			Category:      c.Category,
			Difficulty:    c.Difficulty,
			Points:        c.Points,
			IsSolved:      solvedMap[c.ID],
			SolvedCount:   solvedCountMap[c.ID],
			TotalAttempts: attemptsMap[c.ID],
			CreatedAt:     c.CreatedAt,
		})
	}

	return &dto.PageResult{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

// GetPublishedChallenge 获取已发布靶场详情（学员视图）
func (s *Service) GetPublishedChallenge(userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}

	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errcode.ErrForbidden
	}

	var isSolved bool
	if userID > 0 {
		isSolved, err = s.repo.GetSolvedStatus(userID, challengeID)
		if err != nil {
			s.log.Error("failed to get solved status", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeID), zap.Error(err))
		}
	}

	solvedCount, err := s.getSolvedCountCached(challengeID)
	if err != nil {
		s.log.Warn("failed to get solved count", zap.Int64("challenge_id", challengeID), zap.Error(err))
		solvedCount = 0
	}

	attempts, err := s.repo.GetTotalAttempts(challengeID)
	if err != nil {
		s.log.Error("failed to get total attempts", zap.Int64("challenge_id", challengeID), zap.Error(err))
		attempts = 0
	}

	return &dto.ChallengeDetailResp{
		ID:            challenge.ID,
		Title:         challenge.Title,
		Description:   challenge.Description,
		Category:      challenge.Category,
		Difficulty:    challenge.Difficulty,
		Points:        challenge.Points,
		SolvedCount:   solvedCount,
		TotalAttempts: attempts,
		IsSolved:      isSolved,
		CreatedAt:     challenge.CreatedAt,
	}, nil
}

// getSolvedCountCached 获取完成人数（带缓存）
func (s *Service) getSolvedCountCached(challengeID int64) (int64, error) {
	ctx := context.Background()
	cacheKey := cache.ChallengeSolvedCountKey(challengeID)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var count int64
		if json.Unmarshal([]byte(cached), &count) == nil {
			return count, nil
		}
	} else if err != redis.Nil {
		s.log.Error("redis get failed, fallback to db", zap.String("key", cacheKey), zap.Error(err))
	}

	count, err := s.repo.GetSolvedCount(challengeID)
	if err != nil {
		return 0, err
	}

	data, _ := json.Marshal(count)
	s.redis.Set(ctx, cacheKey, data, s.config.SolvedCountCacheTTL)

	return count, nil
}
