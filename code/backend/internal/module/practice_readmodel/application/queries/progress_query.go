package queries

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *QueryService) GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error) {
	cacheKey := constants.UserProgressKey(userID)
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey).Result()
		if err == nil {
			var resp dto.ProgressResp
			if json.Unmarshal([]byte(cached), &resp) == nil {
				return &resp, nil
			}
			s.logger.Warn("进度缓存反序列化失败", zap.Int64("user_id", userID))
		}
	}

	totalScore, totalSolved, err := s.repo.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rank, err := s.repo.GetUserRank(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	categoryStats, err := s.repo.GetCategoryStats(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	difficultyStats, err := s.repo.GetDifficultyStats(ctx, userID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := &dto.ProgressResp{
		TotalScore:      totalScore,
		TotalSolved:     totalSolved,
		Rank:            rank,
		CategoryStats:   make([]dto.CategoryStat, len(categoryStats)),
		DifficultyStats: make([]dto.DifficultyStat, len(difficultyStats)),
	}
	for i, stat := range categoryStats {
		resp.CategoryStats[i] = dto.CategoryStat{
			Category: stat.Category,
			Solved:   stat.Solved,
			Total:    stat.Total,
		}
	}
	for i, stat := range difficultyStats {
		resp.DifficultyStats[i] = dto.DifficultyStat{
			Difficulty: stat.Difficulty,
			Solved:     stat.Solved,
			Total:      stat.Total,
		}
	}

	if s.cache != nil {
		if data, err := json.Marshal(resp); err == nil {
			_ = s.cache.Set(ctx, cacheKey, data, s.cacheTTL).Err()
		}
	}

	return resp, nil
}
