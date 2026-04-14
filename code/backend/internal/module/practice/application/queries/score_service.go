package queries

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ScoreService struct {
	repo   practiceports.PracticeRankingRepository
	logger *zap.Logger
	config *config.ScoreConfig
}

func NewScoreService(repo practiceports.PracticeRankingRepository, logger *zap.Logger, cfg *config.ScoreConfig) *ScoreService {
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
		repo:   repo,
		logger: logger,
		config: cfg,
	}
}

func (s *ScoreService) GetUserScore(userID int64) (*dto.UserScoreInfo, error) {
	return s.GetUserScoreWithContext(context.Background(), userID)
}

func (s *ScoreService) GetUserScoreWithContext(ctx context.Context, userID int64) (*dto.UserScoreInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	userScore, err := s.repo.FindUserScoreWithContext(ctx, userID)
	if err == gorm.ErrRecordNotFound {
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

	userMap, userErr := s.getUserProfilesWithContext(ctx, []int64{userID})
	if userErr != nil {
		s.logger.Warn("查询用户名失败", zap.Int64("userID", userID), zap.Error(userErr))
	}

	info := &dto.UserScoreInfo{
		UserID:      userScore.UserID,
		Username:    userMap[userID].Username,
		TotalScore:  userScore.TotalScore,
		SolvedCount: userScore.SolvedCount,
		Rank:        userScore.Rank,
	}

	return info, nil
}

func (s *ScoreService) GetRanking(limit int) ([]*dto.RankingItem, error) {
	return s.GetRankingWithContext(context.Background(), limit)
}

func (s *ScoreService) GetRankingWithContext(ctx context.Context, limit int) ([]*dto.RankingItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 || limit > s.config.MaxRankingLimit {
		limit = s.config.MaxRankingLimit
	}

	scores, err := s.repo.ListTopUserScoresWithContext(ctx, limit)
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, len(scores))
	for idx, score := range scores {
		userIDs[idx] = score.UserID
	}
	userMap, err := s.getUserProfilesWithContext(ctx, userIDs)
	if err != nil {
		s.logger.Error("批量查询用户名失败", zap.Error(err))
	}

	result := make([]*dto.RankingItem, 0, len(scores))
	for idx, score := range scores {
		result = append(result, &dto.RankingItem{
			Rank:        idx + 1,
			UserID:      score.UserID,
			Username:    userMap[score.UserID].Username,
			TotalScore:  score.TotalScore,
			SolvedCount: score.SolvedCount,
			ClassName:   userMap[score.UserID].ClassName,
		})
	}

	return result, nil
}

type userProfile struct {
	Username  string
	ClassName string
}

func (s *ScoreService) getUserProfilesWithContext(ctx context.Context, userIDs []int64) (map[int64]userProfile, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(userIDs) == 0 {
		return make(map[int64]userProfile), nil
	}

	users, err := s.repo.FindUsersByIDsWithContext(ctx, userIDs)
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
