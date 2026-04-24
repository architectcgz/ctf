package commands

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type imageCommandContextRepoStub struct {
	createFn            func(ctx context.Context, image *model.Image) error
	findByIDFn          func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn     func(ctx context.Context, name, tag string) (*model.Image, error)
	listFn              func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn            func(ctx context.Context, image *model.Image) error
	deleteWithContextFn func(ctx context.Context, id int64) error
}

func (s *imageCommandContextRepoStub) Create(ctx context.Context, image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(ctx, image)
	}
	return nil
}

func (s *imageCommandContextRepoStub) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *imageCommandContextRepoStub) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagFn != nil {
		return s.findByNameTagFn(ctx, name, tag)
	}
	return nil, nil
}

func (s *imageCommandContextRepoStub) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *imageCommandContextRepoStub) Update(ctx context.Context, image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, image)
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
	countByImageIDFn func(ctx context.Context, imageID int64) (int64, error)
}

func (s *imageUsageContextStub) CountByImageID(ctx context.Context, imageID int64) (int64, error) {
	if s.countByImageIDFn != nil {
		return s.countByImageIDFn(ctx, imageID)
	}
	return 0, nil
}

type imageCommandContextKey string

func TestImageServiceUpdateImagePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := imageCommandContextKey("image-update")
	expectedCtxValue := "ctx-image-update"
	findCalled := false
	updateCalled := false
	repo := &imageCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-image ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Description: "old", Status: model.ImageStatusAvailable}, nil
		},
		updateFn: func(ctx context.Context, image *model.Image) error {
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
	if err := service.UpdateImage(ctx, 9, &dto.UpdateImageReq{Description: "new"}); err != nil {
		t.Fatalf("UpdateImage() error = %v", err)
	}
	if !findCalled || !updateCalled {
		t.Fatalf("expected repository calls, got find=%v update=%v", findCalled, updateCalled)
	}
}

func TestImageServiceDeleteImagePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := imageCommandContextKey("image-delete")
	expectedCtxValue := "ctx-image-delete"
	findCalled := false
	countCalled := false
	deleteCalled := false
	repo := &imageCommandContextRepoStub{
		findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
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
		countByImageIDFn: func(ctx context.Context, imageID int64) (int64, error) {
			countCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected count-usage ctx value %v, got %v", expectedCtxValue, got)
			}
			return 0, nil
		},
	}
	service := NewImageService(repo, usageRepo, nil, zap.NewNop())

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if err := service.DeleteImage(ctx, 9); err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}
	if !findCalled || !countCalled || !deleteCalled {
		t.Fatalf("expected repository calls, got find=%v count=%v delete=%v", findCalled, countCalled, deleteCalled)
	}
}
