package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ScoreService struct {
	repo       practiceRankingRepository
	stateStore practiceports.PracticeScoreStateStore
	logger     *zap.Logger
	config     *config.ScoreConfig
}

type practiceRankingRepository interface {
	practiceports.PracticeUserScoreReadRepository
	practiceports.PracticeRankingListRepository
	practiceports.PracticeUserDirectoryRepository
}

func NewScoreService(repo practiceRankingRepository, stateStore practiceports.PracticeScoreStateStore, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg == nil {
		cfg = &config.ScoreConfig{}
	}
	if cfg.MaxRankingLimit <= 0 {
		cfg.MaxRankingLimit = 100
	}
	return &ScoreService{
		repo:       repo,
		stateStore: stateStore,
		logger:     logger,
		config:     cfg,
	}
}

func (s *ScoreService) GetUserScore(ctx context.Context, userID int64) (*dto.UserScoreInfo, error) {
	if s.stateStore != nil {
		cached, hit, err := s.stateStore.LoadUserScoreCache(ctx, userID)
		if err == nil && hit {
			return cached, nil
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			s.logger.Warn("读取用户得分缓存失败", zap.Int64("userID", userID), zap.Error(err))
		}
	}

	userScore, err := s.repo.FindUserScore(ctx, userID)
	if errors.Is(err, practiceports.ErrPracticeUserScoreNotFound) {
		return &dto.UserScoreInfo{
			UserID:      userID,
			TotalScore:  0,
			SolvedCount: 0,
			Rank:        0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	userProfiles, userErr := s.getUserProfiles(ctx, []int64{userID})
	if userErr != nil {
		s.logger.Warn("查询用户名失败", zap.Int64("userID", userID), zap.Error(userErr))
	}

	info := practiceQueryResponseMapperInst.ToUserScoreInfoBasePtr(userScore)
	info.Username = userProfiles[userID].Username

	if s.stateStore != nil {
		if err := s.stateStore.StoreUserScoreCache(ctx, info, s.cacheTTL()); err != nil && ctx.Err() == nil {
			s.logger.Warn("回填用户得分缓存失败", zap.Int64("userID", userID), zap.Error(err))
		}
	}

	return info, nil
}

func (s *ScoreService) GetRanking(ctx context.Context, limit int) ([]*dto.RankingItem, error) {
	if limit <= 0 || limit > s.config.MaxRankingLimit {
		limit = s.config.MaxRankingLimit
	}
	scores, err := s.repo.ListTopUserScores(ctx, limit)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, len(scores))
	for idx, score := range scores {
		userIDs[idx] = score.UserID
	}
	userProfiles, err := s.getUserProfiles(ctx, userIDs)
	if err != nil {
		s.logger.Error("批量查询用户名失败", zap.Error(err))
	}

	result := make([]*dto.RankingItem, 0, len(scores))
	for idx, score := range scores {
		scoreCopy := score
		item := practiceQueryResponseMapperInst.ToRankingItemBasePtr(&scoreCopy)
		item.Rank = idx + 1
		item.Username = userProfiles[score.UserID].Username
		item.ClassName = userProfiles[score.UserID].ClassName
		result = append(result, item)
	}

	return result, nil
}

type userProfile struct {
	Username  string
	ClassName string
}

func (s *ScoreService) getUserProfiles(ctx context.Context, userIDs []int64) (map[int64]userProfile, error) {
	if len(userIDs) == 0 {
		return make(map[int64]userProfile), nil
	}

	users, err := s.repo.FindUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]userProfile, len(userIDs))
	for _, user := range users {
		result[user.ID] = userProfile{
			Username:  user.Username,
			ClassName: user.ClassName,
		}
	}
	for _, userID := range userIDs {
		if _, exists := result[userID]; exists {
			continue
		}
		result[userID] = userProfile{Username: fmt.Sprintf("用户%d", userID)}
		s.logger.Warn("用户不存在", zap.Int64("userID", userID))
	}

	return result, nil
}

func (s *ScoreService) cacheTTL() time.Duration {
	if s.config == nil || s.config.CacheTTL <= 0 {
		return 0
	}
	return s.config.CacheTTL
}
