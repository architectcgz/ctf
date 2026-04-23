package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type tagCommandContextStub struct {
	createFn                            func(tag *model.Tag) error
	createWithContextFn                 func(ctx context.Context, tag *model.Tag) error
	findByIDsFn                         func(ids []int64) ([]*model.Tag, error)
	findByIDsWithContextFn              func(ctx context.Context, ids []int64) ([]*model.Tag, error)
	attachTagsInTxFn                    func(challengeID int64, tagIDs []int64) error
	attachTagsInTxWithContextFn         func(ctx context.Context, challengeID int64, tagIDs []int64) error
	detachFromChallengeFn               func(challengeID, tagID int64) error
	detachFromChallengeWithContextFn    func(ctx context.Context, challengeID, tagID int64) error
	countChallengesByTagIDFn            func(tagID int64) (int64, error)
	countChallengesByTagIDWithContextFn func(ctx context.Context, tagID int64) (int64, error)
	deleteFn                            func(id int64) error
	deleteWithContextFn                 func(ctx context.Context, id int64) error
}

func (s *tagCommandContextStub) Create(tag *model.Tag) error {
	if s.createFn != nil {
		return s.createFn(tag)
	}
	return nil
}

func (s *tagCommandContextStub) CreateWithContext(ctx context.Context, tag *model.Tag) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, tag)
	}
	return s.Create(tag)
}

func (s *tagCommandContextStub) List(tagType string) ([]*model.Tag, error) { return nil, nil }
func (s *tagCommandContextStub) ListWithContext(ctx context.Context, tagType string) ([]*model.Tag, error) {
	return s.List(tagType)
}

func (s *tagCommandContextStub) FindByIDs(ids []int64) ([]*model.Tag, error) {
	if s.findByIDsFn != nil {
		return s.findByIDsFn(ids)
	}
	return nil, nil
}

func (s *tagCommandContextStub) FindByIDsWithContext(ctx context.Context, ids []int64) ([]*model.Tag, error) {
	if s.findByIDsWithContextFn != nil {
		return s.findByIDsWithContextFn(ctx, ids)
	}
	return s.FindByIDs(ids)
}

func (s *tagCommandContextStub) AttachTagsInTx(challengeID int64, tagIDs []int64) error {
	if s.attachTagsInTxFn != nil {
		return s.attachTagsInTxFn(challengeID, tagIDs)
	}
	return nil
}

func (s *tagCommandContextStub) AttachTagsInTxWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
	if s.attachTagsInTxWithContextFn != nil {
		return s.attachTagsInTxWithContextFn(ctx, challengeID, tagIDs)
	}
	return s.AttachTagsInTx(challengeID, tagIDs)
}

func (s *tagCommandContextStub) DetachFromChallenge(challengeID, tagID int64) error {
	if s.detachFromChallengeFn != nil {
		return s.detachFromChallengeFn(challengeID, tagID)
	}
	return nil
}

func (s *tagCommandContextStub) DetachFromChallengeWithContext(ctx context.Context, challengeID, tagID int64) error {
	if s.detachFromChallengeWithContextFn != nil {
		return s.detachFromChallengeWithContextFn(ctx, challengeID, tagID)
	}
	return s.DetachFromChallenge(challengeID, tagID)
}

func (s *tagCommandContextStub) FindByChallengeID(challengeID int64) ([]*model.Tag, error) {
	return nil, nil
}
func (s *tagCommandContextStub) FindByChallengeIDWithContext(ctx context.Context, challengeID int64) ([]*model.Tag, error) {
	return s.FindByChallengeID(challengeID)
}
func (s *tagCommandContextStub) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}
func (s *tagCommandContextStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return s.Delete(id)
}
func (s *tagCommandContextStub) CountChallengesByTagID(tagID int64) (int64, error) {
	if s.countChallengesByTagIDFn != nil {
		return s.countChallengesByTagIDFn(tagID)
	}
	return 0, nil
}
func (s *tagCommandContextStub) CountChallengesByTagIDWithContext(ctx context.Context, tagID int64) (int64, error) {
	if s.countChallengesByTagIDWithContextFn != nil {
		return s.countChallengesByTagIDWithContextFn(ctx, tagID)
	}
	return s.CountChallengesByTagID(tagID)
}

type tagCommandContextKey string

func TestTagServiceCreateTagWithContextPropagatesContextToRepository(t *testing.T) {
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
	resp, err := service.CreateTagWithContext(ctx, &dto.CreateTagReq{Name: "SQL注入", Type: model.TagTypeVulnerability, Description: "desc"})
	if err != nil {
		t.Fatalf("CreateTagWithContext() error = %v", err)
	}
	if !createCalled {
		t.Fatal("expected create repository to be called")
	}
	if resp == nil || resp.ID != 11 || resp.Name != "SQL注入" {
		t.Fatalf("unexpected tag resp: %+v", resp)
	}
}

func TestTagServiceAttachTagsWithContextPropagatesContextToRepository(t *testing.T) {
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
	if err := service.AttachTagsWithContext(ctx, 99, []int64{1, 2}); err != nil {
		t.Fatalf("AttachTagsWithContext() error = %v", err)
	}
	if !findByIDsCalled || !attachCalled {
		t.Fatalf("expected repository calls, got find=%v attach=%v", findByIDsCalled, attachCalled)
	}
}

func TestTagServiceDetachTagsWithContextPropagatesContextToRepository(t *testing.T) {
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
	if err := service.DetachTagsWithContext(ctx, 99, []int64{1, 2}); err != nil {
		t.Fatalf("DetachTagsWithContext() error = %v", err)
	}
	if !findByIDsCalled || detachCalls != 2 {
		t.Fatalf("expected repository calls, got find=%v detachCalls=%d", findByIDsCalled, detachCalls)
	}
}

func TestTagServiceDeleteTagWithContextPropagatesContextToRepository(t *testing.T) {
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
	if err := service.DeleteTagWithContext(ctx, 11); err != nil {
		t.Fatalf("DeleteTagWithContext() error = %v", err)
	}
	if !countCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got count=%v delete=%v", countCalled, deleteCalled)
	}
}
