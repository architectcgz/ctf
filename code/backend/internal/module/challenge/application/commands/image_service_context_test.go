package commands

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type imageCommandContextRepoStub struct {
	createWithContextFn        func(ctx context.Context, image *model.Image) error
	findByIDWithContextFn      func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagWithContextFn func(ctx context.Context, name, tag string) (*model.Image, error)
	listWithContextFn          func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateWithContextFn        func(ctx context.Context, image *model.Image) error
	deleteWithContextFn        func(ctx context.Context, id int64) error
}

func (s *imageCommandContextRepoStub) CreateWithContext(ctx context.Context, image *model.Image) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, image)
	}
	return nil
}

func (s *imageCommandContextRepoStub) FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *imageCommandContextRepoStub) FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagWithContextFn != nil {
		return s.findByNameTagWithContextFn(ctx, name, tag)
	}
	return nil, nil
}

func (s *imageCommandContextRepoStub) ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *imageCommandContextRepoStub) UpdateWithContext(ctx context.Context, image *model.Image) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, image)
	}
	return nil
}

func (s *imageCommandContextRepoStub) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return nil
}

type imageUsageContextStub struct {
	countByImageIDFn            func(imageID int64) (int64, error)
	countByImageIDWithContextFn func(ctx context.Context, imageID int64) (int64, error)
}

func (s *imageUsageContextStub) CountByImageID(imageID int64) (int64, error) {
	if s.countByImageIDFn != nil {
		return s.countByImageIDFn(imageID)
	}
	return 0, nil
}

func (s *imageUsageContextStub) CountByImageIDWithContext(ctx context.Context, imageID int64) (int64, error) {
	if s.countByImageIDWithContextFn != nil {
		return s.countByImageIDWithContextFn(ctx, imageID)
	}
	return s.CountByImageID(imageID)
}

type imageCommandContextKey string

func TestImageServiceUpdateImageWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := imageCommandContextKey("image-update")
	expectedCtxValue := "ctx-image-update"
	findCalled := false
	updateCalled := false
	repo := &imageCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Description: "old", Status: model.ImageStatusAvailable}, nil
		},
		updateWithContextFn: func(ctx context.Context, image *model.Image) error {
			updateCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-image ctx value %v, got %v", expectedCtxValue, got)
			}
			if image.Description != "new" {
				t.Fatalf("unexpected image payload: %+v", image)
			}
			return nil
		},
	}
	service := NewImageService(repo, &imageUsageContextStub{}, nil, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.UpdateImageWithContext(ctx, 9, &dto.UpdateImageReq{Description: "new"}); err != nil {
		t.Fatalf("UpdateImageWithContext() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
}

func TestImageServiceDeleteImageWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := imageCommandContextKey("image-delete")
	expectedCtxValue := "ctx-image-delete"
	findCalled := false
	countCalled := false
	deleteCalled := false
	repo := &imageCommandContextRepoStub{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable}, nil
		},
		deleteWithContextFn: func(ctx context.Context, id int64) error {
			deleteCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected delete-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	usageRepo := &imageUsageContextStub{
		countByImageIDWithContextFn: func(ctx context.Context, imageID int64) (int64, error) {
			countCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected count-usage ctx value %v, got %v", expectedCtxValue, got)
			}
			return 0, nil
		},
	}
	service := NewImageService(repo, usageRepo, nil, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteImageWithContext(ctx, 9); err != nil {
		t.Fatalf("DeleteImageWithContext() error = %v", err)
	}
	if !findCalled || !countCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v count=%v delete=%v", findCalled, countCalled, deleteCalled)
	}
}
