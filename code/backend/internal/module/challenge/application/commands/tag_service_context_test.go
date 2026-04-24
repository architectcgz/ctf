package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type tagCommandContextStub struct {
	createWithContextFn                 func(ctx context.Context, tag *model.Tag) error
	findByIDsWithContextFn              func(ctx context.Context, ids []int64) ([]*model.Tag, error)
	attachTagsInTxWithContextFn         func(ctx context.Context, challengeID int64, tagIDs []int64) error
	detachFromChallengeWithContextFn    func(ctx context.Context, challengeID, tagID int64) error
	countChallengesByTagIDWithContextFn func(ctx context.Context, tagID int64) (int64, error)
	deleteWithContextFn                 func(ctx context.Context, id int64) error
}

func (s *tagCommandContextStub) CreateWithContext(ctx context.Context, tag *model.Tag) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, tag)
	}
	return nil
}

func (s *tagCommandContextStub) ListWithContext(ctx context.Context, tagType string) ([]*model.Tag, error) {
	return nil, nil
}

func (s *tagCommandContextStub) FindByIDsWithContext(ctx context.Context, ids []int64) ([]*model.Tag, error) {
	if s.findByIDsWithContextFn != nil {
		return s.findByIDsWithContextFn(ctx, ids)
	}
	return nil, nil
}

func (s *tagCommandContextStub) AttachTagsInTxWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
	if s.attachTagsInTxWithContextFn != nil {
		return s.attachTagsInTxWithContextFn(ctx, challengeID, tagIDs)
	}
	return nil
}

func (s *tagCommandContextStub) DetachFromChallengeWithContext(ctx context.Context, challengeID, tagID int64) error {
	if s.detachFromChallengeWithContextFn != nil {
		return s.detachFromChallengeWithContextFn(ctx, challengeID, tagID)
	}
	return nil
}

func (s *tagCommandContextStub) FindByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.Tag, error) {
	return nil, nil
}

func (s *tagCommandContextStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

func (s *tagCommandContextStub) CountChallengesByTagIDWithContext(ctx context.Context, tagID int64) (int64, error) {
	if s.countChallengesByTagIDWithContextFn != nil {
		return s.countChallengesByTagIDWithContextFn(ctx, tagID)
	}
	return 0, nil
}

type tagCommandContextKey string

func TestTagServiceCreateTagPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagCommandContextKey("tag-create")
	expectedCtxValue := "ctx-tag-create"
	createCalled := false
	repo := &tagCommandContextStub{
		createWithContextFn: func(ctx context.Context, tag *model.Tag) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create ctx value %v, got %v", expectedCtxValue, got)
			}
			tag.ID = 11
			return nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.CreateTag(ctx, &dto.CreateTagReq{Name: "SQL注入", Type: model.TagTypeVulnerability, Description: "desc"})
	if err != nil {
		t.Fatalf("CreateTag() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected create repository to be called")
	}
	if resp == nil || resp.ID != 11 || resp.Name != "SQL注入" {
		t.Fatalf("unexpected tag resp: %+v", resp)
	}
}

func TestTagServiceAttachTagsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagCommandContextKey("tag-attach")
	expectedCtxValue := "ctx-tag-attach"
	findByIDsCalled := false
	attachCalled := false
	repo := &tagCommandContextStub{
		findByIDsWithContextFn: func(ctx context.Context, ids []int64) ([]*model.Tag, error) {
			findByIDsCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-ids ctx value %v, got %v", expectedCtxValue, got)
			}
			return []*model.Tag{{ID: ids[0]}, {ID: ids[1]}}, nil
		},
		attachTagsInTxWithContextFn: func(ctx context.Context, challengeID int64, tagIDs []int64) error {
			attachCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected attach ctx value %v, got %v", expectedCtxValue, got)
			}
			if challengeID != 99 || len(tagIDs) != 2 {
				t.Fatalf("unexpected attach args: challenge=%d tags=%v", challengeID, tagIDs)
			}
			return nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.AttachTags(ctx, 99, []int64{1, 2}); err != nil {
		t.Fatalf("AttachTags() error = %v", err)
	}
	if !findByIDsCalled || !attachCalled {
		t.Fatalf("expected repository calls, got find=%v attach=%v", findByIDsCalled, attachCalled)
	}
}

func TestTagServiceDetachTagsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagCommandContextKey("tag-detach")
	expectedCtxValue := "ctx-tag-detach"
	findByIDsCalled := false
	detachCalls := 0
	repo := &tagCommandContextStub{
		findByIDsWithContextFn: func(ctx context.Context, ids []int64) ([]*model.Tag, error) {
			findByIDsCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-ids ctx value %v, got %v", expectedCtxValue, got)
			}
			return []*model.Tag{{ID: ids[0]}, {ID: ids[1]}}, nil
		},
		detachFromChallengeWithContextFn: func(ctx context.Context, challengeID, tagID int64) error {
			detachCalls++
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected detach ctx value %v, got %v", expectedCtxValue, got)
			}
			if challengeID != 99 {
				t.Fatalf("unexpected challenge id: %d", challengeID)
			}
			return nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DetachTags(ctx, 99, []int64{1, 2}); err != nil {
		t.Fatalf("DetachTags() error = %v", err)
	}
	if !findByIDsCalled || detachCalls != 2 {
		t.Fatalf("expected repository calls, got find=%v detachCalls=%d", findByIDsCalled, detachCalls)
	}
}

func TestTagServiceDeleteTagPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := tagCommandContextKey("tag-delete")
	expectedCtxValue := "ctx-tag-delete"
	countCalled := false
	deleteCalled := false
	repo := &tagCommandContextStub{
		countChallengesByTagIDWithContextFn: func(ctx context.Context, tagID int64) (int64, error) {
			countCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected count ctx value %v, got %v", expectedCtxValue, got)
			}
			return 0, nil
		},
		deleteWithContextFn: func(ctx context.Context, id int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewTagService(repo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteTag(ctx, 11); err != nil {
		t.Fatalf("DeleteTag() error = %v", err)
	}
	if !countCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got count=%v delete=%v", countCalled, deleteCalled)
	}
}
