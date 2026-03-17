package challenge

import (
	"context"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/cache"
	"ctf-platform/pkg/errcode"
	"encoding/json"
	"errors"
	"sort"
	"strings"

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
	if req.ImageID > 0 {
		// 验证镜像是否存在
		_, err := s.imageRepo.FindByID(req.ImageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound.WithCause(errors.New(ErrMsgImageNotFound))
			}
			return nil, err
		}
	}

	challenge := &model.Challenge{
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Difficulty:    req.Difficulty,
		Points:        req.Points,
		ImageID:       req.ImageID,
		AttachmentURL: strings.TrimSpace(req.AttachmentURL),
		Status:        model.ChallengeStatusDraft,
	}

	hints, err := normalizeHintModels(req.Hints)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateWithHints(challenge, hints); err != nil {
		return nil, err
	}

	return s.toResp(challenge, hints), nil
}

func (s *Service) UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error {
	challenge, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
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
	if req.ImageID != nil {
		if *req.ImageID > 0 {
			_, err := s.imageRepo.FindByID(*req.ImageID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errcode.ErrNotFound.WithCause(errors.New(ErrMsgImageNotFound))
				}
				return err
			}
		}
		challenge.ImageID = *req.ImageID
	}
	if req.AttachmentURL != nil {
		challenge.AttachmentURL = strings.TrimSpace(*req.AttachmentURL)
	}

	replaceHints := req.Hints != nil
	hints, err := normalizeHintModels(req.Hints)
	if err != nil {
		return err
	}

	return s.repo.UpdateWithHints(challenge, hints, replaceHints)
}

func (s *Service) DeleteChallenge(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	hints, err := s.repo.ListHintsByChallengeID(id)
	if err != nil {
		return nil, err
	}
	return s.toResp(challenge, hints), nil
}

func (s *Service) ListChallenges(query *dto.ChallengeQuery) (*dto.PageResult, error) {
	challenges, total, err := s.repo.List(query)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ChallengeResp, len(challenges))
	for i, c := range challenges {
		list[i] = s.toResp(c, nil)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrChallengeNotFound
		}
		return err
	}

	challenge.Status = model.ChallengeStatusPublished
	return s.repo.Update(challenge)
}

func (s *Service) toResp(c *model.Challenge, hints []*model.ChallengeHint) *dto.ChallengeResp {
	adminHints := make([]*dto.ChallengeHintAdminResp, 0, len(hints))
	for _, hint := range hints {
		adminHints = append(adminHints, &dto.ChallengeHintAdminResp{
			ID:         hint.ID,
			Level:      hint.Level,
			Title:      hint.Title,
			CostPoints: hint.CostPoints,
			Content:    hint.Content,
		})
	}

	return &dto.ChallengeResp{
		ID:            c.ID,
		Title:         c.Title,
		Description:   c.Description,
		Category:      c.Category,
		Difficulty:    c.Difficulty,
		Points:        c.Points,
		ImageID:       c.ImageID,
		AttachmentURL: c.AttachmentURL,
		Hints:         adminHints,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// ListPublishedChallenges 查询已发布靶场列表（学员视图）
func (s *Service) ListPublishedChallenges(userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	return s.ListPublishedChallengesWithContext(context.Background(), userID, query)
}

func (s *Service) ListPublishedChallengesWithContext(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	challenges, total, err := s.repo.ListPublishedWithContext(ctx, query)
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
		solvedMap, err = s.repo.BatchGetSolvedStatusWithContext(ctx, userID, challengeIDs)
		if err != nil {
			s.log.Error("failed to batch get solved status", zap.Error(err))
		}
	}

	solvedCountMap, err := s.repo.BatchGetSolvedCountWithContext(ctx, challengeIDs)
	if err != nil {
		s.log.Error("failed to batch get solved count", zap.Error(err))
	}

	attemptsMap, err := s.repo.BatchGetTotalAttemptsWithContext(ctx, challengeIDs)
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
	return s.GetPublishedChallengeWithContext(context.Background(), userID, challengeID)
}

func (s *Service) GetPublishedChallengeWithContext(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	challenge, err := s.repo.FindByIDWithContext(ctx, challengeID)
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
		isSolved, err = s.repo.GetSolvedStatusWithContext(ctx, userID, challengeID)
		if err != nil {
			s.log.Error("failed to get solved status", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeID), zap.Error(err))
		}
	}

	solvedCount, err := s.getSolvedCountCached(ctx, challengeID)
	if err != nil {
		s.log.Warn("failed to get solved count", zap.Int64("challenge_id", challengeID), zap.Error(err))
		solvedCount = 0
	}

	attempts, err := s.repo.GetTotalAttemptsWithContext(ctx, challengeID)
	if err != nil {
		s.log.Error("failed to get total attempts", zap.Int64("challenge_id", challengeID), zap.Error(err))
		attempts = 0
	}

	hints, err := s.repo.ListHintsByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	unlockedHintIDs, err := s.repo.GetUnlockedHintIDsWithContext(ctx, userID, challengeID)
	if err != nil {
		return nil, err
	}

	hintList := make([]*dto.ChallengeHintResp, 0, len(hints))
	for _, hint := range hints {
		hintResp := &dto.ChallengeHintResp{
			ID:         hint.ID,
			Level:      hint.Level,
			Title:      hint.Title,
			CostPoints: hint.CostPoints,
			IsUnlocked: unlockedHintIDs[hint.ID],
		}
		if hintResp.IsUnlocked {
			hintResp.Content = hint.Content
		}
		hintList = append(hintList, hintResp)
	}

	return &dto.ChallengeDetailResp{
		ID:            challenge.ID,
		Title:         challenge.Title,
		Description:   challenge.Description,
		Category:      challenge.Category,
		Difficulty:    challenge.Difficulty,
		Points:        challenge.Points,
		NeedTarget:    challenge.ImageID > 0,
		AttachmentURL: challenge.AttachmentURL,
		Hints:         hintList,
		SolvedCount:   solvedCount,
		TotalAttempts: attempts,
		IsSolved:      isSolved,
		CreatedAt:     challenge.CreatedAt,
	}, nil
}

// getSolvedCountCached 获取完成人数（带缓存）
func (s *Service) getSolvedCountCached(ctx context.Context, challengeID int64) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if s.redis == nil {
		return s.repo.GetSolvedCountWithContext(ctx, challengeID)
	}

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

	count, err := s.repo.GetSolvedCountWithContext(ctx, challengeID)
	if err != nil {
		return 0, err
	}

	data, _ := json.Marshal(count)
	s.redis.Set(ctx, cacheKey, data, s.config.SolvedCountCacheTTL)

	return count, nil
}

func normalizeHintModels(reqHints []dto.ChallengeHintReq) ([]*model.ChallengeHint, error) {
	if reqHints == nil {
		return nil, nil
	}

	hints := make([]*model.ChallengeHint, 0, len(reqHints))
	seenLevels := make(map[int]struct{}, len(reqHints))
	for _, reqHint := range reqHints {
		content := strings.TrimSpace(reqHint.Content)
		if content == "" {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("提示内容不能为空"))
		}
		if _, exists := seenLevels[reqHint.Level]; exists {
			return nil, errcode.ErrInvalidParams.WithCause(errors.New("提示级别不能重复"))
		}
		seenLevels[reqHint.Level] = struct{}{}
		hints = append(hints, &model.ChallengeHint{
			Level:      reqHint.Level,
			Title:      strings.TrimSpace(reqHint.Title),
			CostPoints: reqHint.CostPoints,
			Content:    content,
		})
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i].Level < hints[j].Level
	})
	return hints, nil
}
