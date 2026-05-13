package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type scoreQuerySourceStub struct {
	findUserScoreFn     func(context.Context, int64) (*model.UserScore, error)
	listTopUserScoresFn func(context.Context, int) ([]model.UserScore, error)
	findUsersByIDsFn    func(context.Context, []int64) ([]model.User, error)
}

func (s scoreQuerySourceStub) FindUserScore(ctx context.Context, userID int64) (*model.UserScore, error) {
	return s.findUserScoreFn(ctx, userID)
}

func (s scoreQuerySourceStub) ListTopUserScores(ctx context.Context, limit int) ([]model.UserScore, error) {
	if s.listTopUserScoresFn == nil {
		return []model.UserScore{}, nil
	}
	return s.listTopUserScoresFn(ctx, limit)
}

func (s scoreQuerySourceStub) FindUsersByIDs(ctx context.Context, userIDs []int64) ([]model.User, error) {
	if s.findUsersByIDsFn == nil {
		return []model.User{}, nil
	}
	return s.findUsersByIDsFn(ctx, userIDs)
}

func TestScoreQueryRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewScoreQueryRepository(scoreQuerySourceStub{
		findUserScoreFn: func(context.Context, int64) (*model.UserScore, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindUserScore(context.Background(), 1); !errors.Is(err, practiceports.ErrPracticeUserScoreNotFound) {
		t.Fatalf("error = %v, want %v", err, practiceports.ErrPracticeUserScoreNotFound)
	}
}

func TestScoreQueryRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewScoreQueryRepository(scoreQuerySourceStub{
		findUserScoreFn: func(context.Context, int64) (*model.UserScore, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindUserScore(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
