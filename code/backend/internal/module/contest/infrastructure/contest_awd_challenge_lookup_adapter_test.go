package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

type contestAWDChallengeLookupSourceStub struct {
	findByIDFn func(context.Context, int64) (*model.AWDChallenge, error)
}

func (s contestAWDChallengeLookupSourceStub) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.AWDChallenge{ID: id}, nil
}

func (s contestAWDChallengeLookupSourceStub) ListAWDChallenges(context.Context, *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return nil, 0, nil
}

func TestContestAWDChallengeLookupAdapterMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
	}{
		{name: "gorm", err: gorm.ErrRecordNotFound},
		{name: "challenge sentinel", err: challengeports.ErrAWDChallengeNotFound},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := NewContestAWDChallengeLookupAdapter(contestAWDChallengeLookupSourceStub{
				findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
					return nil, tc.err
				},
			})

			if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDChallengeNotFound) {
				t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDChallengeNotFound)
			}
		})
	}
}

func TestContestAWDChallengeLookupAdapterPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("awd challenge lookup exploded")
	repo := NewContestAWDChallengeLookupAdapter(contestAWDChallengeLookupSourceStub{
		findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
