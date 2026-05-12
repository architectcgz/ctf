package queries

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *ProgressTimelineService) GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error) {
	if s.cache != nil {
		cached, hit, err := s.cache.GetUserProgress(ctx, userID)
		if err != nil {
			s.logger.Warn("读取用户进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		} else if hit {
			return cached, nil
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
		if err := s.cache.StoreUserProgress(ctx, userID, resp, s.cacheTTL); err != nil {
			s.logger.Warn("保存用户进度缓存失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	return resp, nil
}
