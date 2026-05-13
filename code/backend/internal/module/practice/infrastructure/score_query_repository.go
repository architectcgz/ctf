package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type scoreQuerySource interface {
	practiceports.PracticeUserScoreReadRepository
	practiceports.PracticeRankingListRepository
	practiceports.PracticeUserDirectoryRepository
}

type ScoreQueryRepository struct {
	source scoreQuerySource
}

func NewScoreQueryRepository(source scoreQuerySource) *ScoreQueryRepository {
	if source == nil {
		return nil
	}
	return &ScoreQueryRepository{source: source}
}

func (r *ScoreQueryRepository) FindUserScore(ctx context.Context, userID int64) (*model.UserScore, error) {
	userScore, err := r.source.FindUserScore(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeUserScoreNotFound
	}
	return userScore, err
}

func (r *ScoreQueryRepository) ListTopUserScores(ctx context.Context, limit int) ([]model.UserScore, error) {
	return r.source.ListTopUserScores(ctx, limit)
}

func (r *ScoreQueryRepository) FindUsersByIDs(ctx context.Context, userIDs []int64) ([]model.User, error) {
	return r.source.FindUsersByIDs(ctx, userIDs)
}

var _ interface {
	practiceports.PracticeUserScoreReadRepository
	practiceports.PracticeRankingListRepository
	practiceports.PracticeUserDirectoryRepository
} = (*ScoreQueryRepository)(nil)
