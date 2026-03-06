package challenge

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	imageRepo *ImageRepository
	redis     *redis.Client
}

func NewService(repo *Repository, imageRepo *ImageRepository, redis *redis.Client) *Service {
	return &Service{
		repo:      repo,
		imageRepo: imageRepo,
		redis:     redis,
	}
}

func (s *Service) CreateChallenge(req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	// 验证镜像是否存在
	_, err := s.imageRepo.FindByID(req.ImageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("镜像不存在")
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
			return errors.New("镜像不存在")
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
		return errors.New("存在运行中的实例，无法删除")
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
		return errors.New("靶场未关联镜像，无法发布")
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

	list := make([]*dto.ChallengeListItem, 0, len(challenges))
	for _, c := range challenges {
		item := &dto.ChallengeListItem{
			ID:         c.ID,
			Title:      c.Title,
			Category:   c.Category,
			Difficulty: c.Difficulty,
			Points:     c.Points,
			CreatedAt:  c.CreatedAt,
		}

		isSolved, _ := s.repo.GetSolvedStatus(userID, c.ID)
		item.IsSolved = isSolved

		solvedCount, _ := s.getSolvedCountCached(c.ID)
		item.SolvedCount = solvedCount

		attempts, _ := s.repo.GetTotalAttempts(c.ID)
		item.TotalAttempts = attempts

		list = append(list, item)
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

// GetPublishedChallenge 获取已发布靶场详情（学员视图）
func (s *Service) GetPublishedChallenge(userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
	challenge, err := s.repo.FindByID(challengeID)
	if err != nil {
		return nil, err
	}

	if challenge.Status != model.ChallengeStatusPublished {
		return nil, errors.New("challenge not published")
	}

	isSolved, _ := s.repo.GetSolvedStatus(userID, challengeID)
	solvedCount, _ := s.getSolvedCountCached(challengeID)
	attempts, _ := s.repo.GetTotalAttempts(challengeID)

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
		FlagType:      challenge.FlagType,
		CreatedAt:     challenge.CreatedAt,
	}, nil
}

// getSolvedCountCached 获取完成人数（带缓存）
func (s *Service) getSolvedCountCached(challengeID int64) (int64, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("challenge:solved_count:%d", challengeID)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var count int64
		if json.Unmarshal([]byte(cached), &count) == nil {
			return count, nil
		}
	}

	count, err := s.repo.GetSolvedCount(challengeID)
	if err != nil {
		return 0, err
	}

	data, _ := json.Marshal(count)
	s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	return count, nil
}
