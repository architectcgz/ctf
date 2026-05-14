package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

type contestChallengeLookupSourceStub struct {
	findByIDFn             func(context.Context, int64) (*model.Challenge, error)
	batchGetSolvedStatusFn func(context.Context, int64, []int64) (map[int64]bool, error)
	batchGetSolvedCountFn  func(context.Context, []int64) (map[int64]int64, error)
}

func (s contestChallengeLookupSourceStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Challenge{ID: id}, nil
}

func (s contestChallengeLookupSourceStub) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	if s.batchGetSolvedStatusFn != nil {
		return s.batchGetSolvedStatusFn(ctx, userID, challengeIDs)
	}
	return map[int64]bool{}, nil
}

func (s contestChallengeLookupSourceStub) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	if s.batchGetSolvedCountFn != nil {
		return s.batchGetSolvedCountFn(ctx, challengeIDs)
	}
	return map[int64]int64{}, nil
}

func TestContestChallengeLookupAdapterMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
	}{
		{name: "gorm", err: gorm.ErrRecordNotFound},
		{name: "query sentinel", err: challengeports.ErrChallengeQueryChallengeNotFound},
		{name: "command sentinel", err: challengeports.ErrChallengeCommandChallengeNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewContestChallengeLookupAdapter(contestChallengeLookupSourceStub{
				findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
					return nil, tc.err
				},
			})

			if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestChallengeEntityNotFound) {
				t.Fatalf("error = %v, want %v", err, contestports.ErrContestChallengeEntityNotFound)
			}
		})
	}
}

func TestContestChallengeLookupAdapterPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("challenge lookup exploded")
	repo := NewContestChallengeLookupAdapter(contestChallengeLookupSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
