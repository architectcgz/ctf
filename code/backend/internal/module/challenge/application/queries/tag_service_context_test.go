package queries

import (
	"context"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type tagQueryContextStub struct {
	listFn              func(ctx context.Context, tagType string) ([]*challengeentity.Tag, error)
	findByChallengeIDFn func(ctx context.Context, challengeID int64) ([]*challengeentity.Tag, error)
}

func (s *tagQueryContextStub) Create(ctx context.Context, tag *challengeentity.Tag) error {
	return nil
}
func (s *tagQueryContextStub) List(ctx context.Context, tagType string) ([]*challengeentity.Tag, error) {
	if s.listFn != nil {
		return s.listFn(ctx, tagType)
	}
	return nil, nil
}
func (s *tagQueryContextStub) FindByIDs(ctx context.Context, ids []int64) ([]*challengeentity.Tag, error) {
	return nil, nil
}
func (s *tagQueryContextStub) AttachTagsInTx(ctx context.Context, challengeID int64, tagIDs []int64) error {
	return nil
}
func (s *tagQueryContextStub) DetachFromChallenge(ctx context.Context, challengeID, tagID int64) error {
	return nil
}
func (s *tagQueryContextStub) FindByChallengeID(ctx context.Context, challengeID int64) ([]*challengeentity.Tag, error) {
	if s.findByChallengeIDFn != nil {
		return s.findByChallengeIDFn(ctx, challengeID)
	}
	return nil, nil
}
func (s *tagQueryContextStub) Delete(ctx context.Context, id int64) error { return nil }
func (s *tagQueryContextStub) CountChallengesByTagID(ctx context.Context, tagID int64) (int64, error) {
	return 0, nil
}

type tagQueryContextKey string

func TestTagServiceListTagsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagQueryContextKey("tag-list")
	expectedCtxValue := "ctx-tag-list"
	listCalled := false
	repo := &tagQueryContextStub{
		listFn: func(ctx context.Context, tagType string) ([]*challengeentity.Tag, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list ctx value %v, got %v", expectedCtxValue, got)
			}
			if tagType != challengecontracts.TagTypeVulnerability {
				t.Fatalf("unexpected tag type: %s", tagType)
			}
			return []*challengeentity.Tag{{ID: 1, Name: "SQL注入", Type: challengecontracts.TagTypeVulnerability}}, nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListTags(ctx, challengecontracts.TagTypeVulnerability)
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list repository to be called")
	}
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected tag list resp: %+v", resp)
	}
}

func TestTagServiceGetChallengeTagIDsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagQueryContextKey("tag-ids")
	expectedCtxValue := "ctx-tag-ids"
	findCalled := false
	repo := &tagQueryContextStub{
		findByChallengeIDFn: func(ctx context.Context, challengeID int64) ([]*challengeentity.Tag, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-challenge ctx value %v, got %v", expectedCtxValue, got)
			}
			return []*challengeentity.Tag{{ID: 2}, {ID: 5}}, nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetChallengeTagIDs(ctx, 99)
	if err != nil {
		t.Fatalf("GetChallengeTagIDs() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected find-by-challenge repository to be called")
	}
	if len(resp) != 2 || resp[0] != 2 || resp[1] != 5 {
		t.Fatalf("unexpected tag ids resp: %+v", resp)
	}
}
