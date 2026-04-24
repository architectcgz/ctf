package queries

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/cache"
	"ctf-platform/pkg/errcode"
)

type ChallengeService struct {
	repo   challengeports.ChallengeQueryRepository
	redis  *redis.Client
	config *Config
	log    *zap.Logger
}

func NewChallengeService(repo challengeports.ChallengeQueryRepository, redis *redis.Client, config *Config, log *zap.Logger) *ChallengeService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ChallengeService{
		repo:   repo,
		redis:  redis,
		config: config,
		log:    log,
	}
}

func (s *ChallengeService) GetChallenge(ctx context.Context, id int64) (*dto.ChallengeResp, error) {
	challenge, err := s.repo.FindByIDWithContext(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	hints, err := s.repo.ListHintsByChallengeIDWithContext(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.ChallengeRespFromModel(challenge, hints), nil
}

func (s *ChallengeService) ListChallenges(ctx context.Context, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	challenges, total, err := s.repo.ListWithContext(ctx, query)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ChallengeResp, len(challenges))
	for i, challenge := range challenges {
		list[i] = domain.ChallengeRespFromModel(challenge, nil)
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

func (s *ChallengeService) ListPublishedChallenges(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
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

	challengeIDs := make([]int64, len(challenges))
	for index, challenge := range challenges {
		challengeIDs[index] = challenge.ID
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
	for _, challenge := range challenges {
		list = append(list, &dto.ChallengeListItem{
			ID:            challenge.ID,
			Title:         challenge.Title,
			Category:      challenge.Category,
			Difficulty:    challenge.Difficulty,
			Points:        challenge.Points,
			IsSolved:      solvedMap[challenge.ID],
			SolvedCount:   solvedCountMap[challenge.ID],
			TotalAttempts: attemptsMap[challenge.ID],
			CreatedAt:     challenge.CreatedAt,
		})
	}

	return &dto.PageResult{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

func (s *ChallengeService) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
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
		isSolved, err = s.repo.GetSolvedStatus(ctx, userID, challengeID)
		if err != nil {
			s.log.Error("failed to get solved status", zap.Int64("user_id", userID), zap.Int64("challenge_id", challengeID), zap.Error(err))
		}
	}

	solvedCount, err := s.getSolvedCountCached(ctx, challengeID)
	if err != nil {
		s.log.Warn("failed to get solved count", zap.Int64("challenge_id", challengeID), zap.Error(err))
		solvedCount = 0
	}

	attempts, err := s.repo.GetTotalAttempts(ctx, challengeID)
	if err != nil {
		s.log.Error("failed to get total attempts", zap.Int64("challenge_id", challengeID), zap.Error(err))
		attempts = 0
	}

	hints, err := s.repo.ListHintsByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	hintList := make([]*dto.ChallengeHintResp, 0, len(hints))
	for _, hint := range hints {
		hintResp := &dto.ChallengeHintResp{
			ID:      hint.ID,
			Level:   hint.Level,
			Title:   hint.Title,
			Content: hint.Content,
		}
		hintList = append(hintList, hintResp)
	}

	return &dto.ChallengeDetailResp{
		ID:              challenge.ID,
		Title:           challenge.Title,
		Description:     challenge.Description,
		Category:        challenge.Category,
		Difficulty:      challenge.Difficulty,
		Points:          challenge.Points,
		NeedTarget:      challenge.ImageID > 0,
		FlagType:        challenge.FlagType,
		InstanceSharing: challenge.InstanceSharing,
		AttachmentURL:   challenge.AttachmentURL,
		Hints:           hintList,
		SolvedCount:     solvedCount,
		TotalAttempts:   attempts,
		IsSolved:        isSolved,
		CreatedAt:       challenge.CreatedAt,
	}, nil
}

func (s *ChallengeService) getSolvedCountCached(ctx context.Context, challengeID int64) (int64, error) {
	if s.redis == nil {
		return s.repo.GetSolvedCount(ctx, challengeID)
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

	count, err := s.repo.GetSolvedCount(ctx, challengeID)
	if err != nil {
		return 0, err
	}

	data, _ := json.Marshal(count)
	s.redis.Set(ctx, cacheKey, data, s.config.SolvedCountCacheTTL)
	return count, nil
}
