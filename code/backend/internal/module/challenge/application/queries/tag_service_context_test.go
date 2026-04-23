package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/model"
)

type tagQueryContextStub struct {
	listFn              func(tagType string) ([]*model.Tag, error)
	listWithContextFn   func(ctx context.Context, tagType string) ([]*model.Tag, error)
	findByChallengeIDFn func(challengeID int64) ([]*model.Tag, error)
}

func (s *tagQueryContextStub) Create(tag *model.Tag) error { return nil }
func (s *tagQueryContextStub) CreateWithContext(ctx context.Context, tag *model.Tag) error {
	return nil
}
func (s *tagQueryContextStub) List(tagType string) ([]*model.Tag, error) {
	if s.listFn != nil {
		return s.listFn(tagType)
	}
	return nil, nil
}
func (s *tagQueryContextStub) ListWithContext(ctx context.Context, tagType string) ([]*model.Tag, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, tagType)
	}
	return s.List(tagType)
}
func (s *tagQueryContextStub) FindByIDs(ids []int64) ([]*model.Tag, error) { return nil, nil }
func (s *tagQueryContextStub) FindByIDsWithContext(ctx context.Context, ids []int64) ([]*model.Tag, error) {
	return nil, nil
}
func (s *tagQueryContextStub) AttachTagsInTx(challengeID int64, tagIDs []int64) error { return nil }
func (s *tagQueryContextStub) AttachTagsInTxWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
	return nil
}
func (s *tagQueryContextStub) DetachFromChallenge(challengeID, tagID int64) error { return nil }
func (s *tagQueryContextStub) DetachFromChallengeWithContext(ctx context.Context, challengeID, tagID int64) error {
	return nil
}
func (s *tagQueryContextStub) FindByChallengeID(challengeID int64) ([]*model.Tag, error) {
	if s.findByChallengeIDFn != nil {
		return s.findByChallengeIDFn(challengeID)
	}
	return nil, nil
}
func (s *tagQueryContextStub) Delete(id int64) error                             { return nil }
func (s *tagQueryContextStub) CountChallengesByTagID(tagID int64) (int64, error) { return 0, nil }

type tagQueryContextKey string

func TestTagServiceListTagsWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagQueryContextKey("tag-list")
	expectedCtxValue := "ctx-tag-list"
	listCalled := false
	repo := &tagQueryContextStub{
		listWithContextFn: func(ctx context.Context, tagType string) ([]*model.Tag, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list ctx value %v, got %v", expectedCtxValue, got)
			}
			if tagType != model.TagTypeVulnerability {
				t.Fatalf("unexpected tag type: %s", tagType)
			}
			return []*model.Tag{{ID: 1, Name: "SQL注入", Type: model.TagTypeVulnerability}}, nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListTagsWithContext(ctx, model.TagTypeVulnerability)
	if err != nil {
		t.Fatalf("ListTagsWithContext() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list repository to be called")
	}
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Fatalf("unexpected tag list resp: %+v", resp)
	}
}
