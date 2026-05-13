package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeQueryRepositorySourceStub struct {
	findByIDFn          func(context.Context, int64) (*model.Challenge, error)
	listFn              func(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	listPublishedFn     func(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error)
	listHintsFn         func(context.Context, int64) ([]*model.ChallengeHint, error)
	getSolvedStatusFn   func(context.Context, int64, int64) (bool, error)
	getSolvedCountFn    func(context.Context, int64) (int64, error)
	getTotalAttemptsFn  func(context.Context, int64) (int64, error)
	batchSolvedStatusFn func(context.Context, int64, []int64) (map[int64]bool, error)
	batchSolvedCountFn  func(context.Context, []int64) (map[int64]int64, error)
	batchAttemptsFn     func(context.Context, []int64) (map[int64]int64, error)
}

func (s challengeQueryRepositorySourceStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s challengeQueryRepositorySourceStub) List(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return nil, 0, nil
}

func (s challengeQueryRepositorySourceStub) ListPublished(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	if s.listPublishedFn != nil {
		return s.listPublishedFn(ctx, query)
	}
	return nil, 0, nil
}

func (s challengeQueryRepositorySourceStub) ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error) {
	if s.listHintsFn != nil {
		return s.listHintsFn(ctx, challengeID)
	}
	return nil, nil
}

func (s challengeQueryRepositorySourceStub) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	if s.getSolvedStatusFn != nil {
		return s.getSolvedStatusFn(ctx, userID, challengeID)
	}
	return false, nil
}

func (s challengeQueryRepositorySourceStub) GetSolvedCount(ctx context.Context, challengeID int64) (int64, error) {
	if s.getSolvedCountFn != nil {
		return s.getSolvedCountFn(ctx, challengeID)
	}
	return 0, nil
}

func (s challengeQueryRepositorySourceStub) GetTotalAttempts(ctx context.Context, challengeID int64) (int64, error) {
	if s.getTotalAttemptsFn != nil {
		return s.getTotalAttemptsFn(ctx, challengeID)
	}
	return 0, nil
}

func (s challengeQueryRepositorySourceStub) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	if s.batchSolvedStatusFn != nil {
		return s.batchSolvedStatusFn(ctx, userID, challengeIDs)
	}
	return map[int64]bool{}, nil
}

func (s challengeQueryRepositorySourceStub) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	if s.batchSolvedCountFn != nil {
		return s.batchSolvedCountFn(ctx, challengeIDs)
	}
	return map[int64]int64{}, nil
}

func (s challengeQueryRepositorySourceStub) BatchGetTotalAttempts(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	if s.batchAttemptsFn != nil {
		return s.batchAttemptsFn(ctx, challengeIDs)
	}
	return map[int64]int64{}, nil
}

func TestChallengeQueryRepositoryMapsChallengeLookupNotFound(t *testing.T) {
	t.Parallel()

	repo := NewChallengeQueryRepository(challengeQueryRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, challengeports.ErrChallengeQueryChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, challengeports.ErrChallengeQueryChallengeNotFound)
	}
}

func TestChallengeQueryRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewChallengeQueryRepository(challengeQueryRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
