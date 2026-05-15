package queries

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type ChallengeService struct {
	repo             challengeQueryRepository
	solvedCountCache challengeports.ChallengeSolvedCountCache
	config           *Config
	log              *zap.Logger
}

type challengeQueryRepository interface {
	challengeports.ChallengeReadRepository
	challengeports.ChallengePublishedRepository
	challengeports.ChallengeStatsRepository
	challengeports.ChallengeBatchStatsRepository
}

func NewChallengeService(repo challengeQueryRepository, solvedCountCache challengeports.ChallengeSolvedCountCache, config *Config, log *zap.Logger) *ChallengeService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ChallengeService{
		repo:             repo,
		solvedCountCache: solvedCountCache,
		config:           config,
		log:              log,
	}
}

func (s *ChallengeService) GetChallenge(ctx context.Context, id int64) (*dto.ChallengeResp, error) {
	challenge, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeQueryChallengeNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, err
	}
	hints, err := s.repo.ListHintsByChallengeID(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.ChallengeRespFromModel(challenge, hints), nil
}

func (s *ChallengeService) ListChallenges(ctx context.Context, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeResp], error) {
	challenges, total, err := s.repo.List(ctx, query)
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

	return &dto.PageResult[*dto.ChallengeResp]{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *ChallengeService) ListPublishedChallenges(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeListItem], error) {
	challenges, total, err := s.repo.ListPublished(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(challenges) == 0 {
		return &dto.PageResult[*dto.ChallengeListItem]{
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
		solvedMap, err = s.repo.BatchGetSolvedStatus(ctx, userID, challengeIDs)
		if err != nil {
			s.log.Error("failed to batch get solved status", zap.Error(err))
		}
	}

	solvedCountMap, err := s.repo.BatchGetSolvedCount(ctx, challengeIDs)
	if err != nil {
		s.log.Error("failed to batch get solved count", zap.Error(err))
	}

	attemptsMap, err := s.repo.BatchGetTotalAttempts(ctx, challengeIDs)
	if err != nil {
		s.log.Error("failed to batch get total attempts", zap.Error(err))
	}

	list := make([]*dto.ChallengeListItem, 0, len(challenges))
	for _, challenge := range challenges {
		item := challengeQueryResponseMapperInst.ToChallengeListItemBasePtr(challenge)
		item.IsSolved = solvedMap[challenge.ID]
		item.SolvedCount = solvedCountMap[challenge.ID]
		item.TotalAttempts = attemptsMap[challenge.ID]
		list = append(list, item)
	}

	return &dto.PageResult[*dto.ChallengeListItem]{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

func (s *ChallengeService) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
	challenge, err := s.repo.FindByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeQueryChallengeNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, err
	}
	if challenge.Status != model.ChallengeStatusPublished {
		return nil, buildChallengeAccessUnavailableError(challenge.Status)
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

	hints, err := s.repo.ListHintsByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	resp := challengeQueryResponseMapperInst.ToChallengeDetailRespBasePtr(challenge)
	resp.NeedTarget = challenge.ImageID > 0
	resp.Hints = challengeQueryResponseMapperInst.ToChallengeHintRespsPtr(hints)
	resp.SolvedCount = solvedCount
	resp.TotalAttempts = attempts
	resp.IsSolved = isSolved
	return resp, nil
}

func buildChallengeAccessUnavailableError(status model.ChallengeStatus) error {
	switch status {
	case model.ChallengeStatusDraft:
		return errcode.New(
			errcode.ErrChallengeNotPublish.Code,
			"题目为草稿，无法访问",
			errcode.ErrChallengeNotPublish.HTTPStatus,
		)
	case model.ChallengeStatusArchived:
		return errcode.New(
			errcode.ErrChallengeNotPublish.Code,
			"题目已归档，无法访问",
			errcode.ErrChallengeNotPublish.HTTPStatus,
		)
	default:
		return errcode.ErrChallengeNotPublish
	}
}

func (s *ChallengeService) getSolvedCountCached(ctx context.Context, challengeID int64) (int64, error) {
	if s.solvedCountCache == nil {
		return s.repo.GetSolvedCount(ctx, challengeID)
	}

	count, hit, err := s.solvedCountCache.GetSolvedCount(ctx, challengeID)
	if err == nil && hit {
		return count, nil
	}
	if err != nil {
		s.log.Error("solved count cache get failed, fallback to db", zap.Int64("challenge_id", challengeID), zap.Error(err))
	}

	count, err = s.repo.GetSolvedCount(ctx, challengeID)
	if err != nil {
		return 0, err
	}

	if s.config != nil && s.config.SolvedCountCacheTTL > 0 {
		_ = s.solvedCountCache.StoreSolvedCount(ctx, challengeID, count, s.config.SolvedCountCacheTTL)
	}
	return count, nil
}
