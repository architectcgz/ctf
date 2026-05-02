package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type awdChallengeQueryContextRepoStub struct {
	findByIDFn func(ctx context.Context, id int64) (*model.AWDChallenge, error)
	listFn     func(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error)
}

func (s *awdChallengeQueryContextRepoStub) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *awdChallengeQueryContextRepoStub) ListAWDChallenges(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return nil, 0, nil
}

type awdChallengeQueryContextKey string

func TestAWDChallengeQueryServiceGetChallengePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdChallengeQueryContextKey("get")
	expectedCtxValue := "ctx-get"
	findCalled := false
	repo := &awdChallengeQueryContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.AWDChallenge, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.AWDChallenge{ID: id, Name: "Bank Portal AWD", Slug: "bank-portal-awd"}, nil
		},
	}
	service := NewAWDChallengeQueryService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetChallenge(ctx, 77)
	if err != nil {
		t.Fatalf("GetChallenge() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected find repository to be called")
	}
	if resp == nil || resp.ID != 77 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAWDChallengeQueryServiceListChallengesPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := awdChallengeQueryContextKey("list")
	expectedCtxValue := "ctx-list"
	listCalled := false
	repo := &awdChallengeQueryContextRepoStub{
		listFn: func(ctx context.Context, query *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list ctx value %v, got %v", expectedCtxValue, got)
			}
			if query == nil || query.Page != 1 || query.Size != 10 {
				t.Fatalf("unexpected query: %+v", query)
			}
			return []*model.AWDChallenge{
				{ID: 1, Name: "Bank Portal AWD", Slug: "bank-portal-awd"},
			}, 1, nil
		},
	}
	service := NewAWDChallengeQueryService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListChallenges(ctx, ListAWDChallengesInput{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListChallenges() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list repository to be called")
	}
	if resp == nil || resp.Total != 1 || len(resp.Items) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
