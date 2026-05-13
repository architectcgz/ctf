package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
	"errors"
)

type stubChallengeImageRepository struct {
	createFn        func(ctx context.Context, image *model.Image) error
	findByIDFn      func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn func(ctx context.Context, name, tag string) (*model.Image, error)
	listFn          func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn        func(ctx context.Context, image *model.Image) error
	deleteFn        func(ctx context.Context, id int64) error
}

func (s *stubChallengeImageRepository) Create(ctx context.Context, image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(ctx, image)
	}
	return nil
}

func (s *stubChallengeImageRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (s *stubChallengeImageRepository) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagFn != nil {
		return s.findByNameTagFn(ctx, name, tag)
	}
	return nil, nil
}

func (s *stubChallengeImageRepository) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *stubChallengeImageRepository) Update(ctx context.Context, image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, image)
	}
	return nil
}

func (s *stubChallengeImageRepository) Delete(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

type challengeImageContextKey string

func TestImageServiceGetImagePropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeImageContextKey("image-query")
	expectedCtxValue := "ctx-image-query"
	findCalled := false
	repo := &stubChallengeImageRepository{
		findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-id ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable}, nil
		},
	}
	service := NewImageService(repo, &config.Config{Pagination: config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100}})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetImage(ctx, 42)
	if err != nil {
		t.Fatalf("GetImage() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected repository find to be called")
	}
	if resp == nil || resp.ID != 42 || resp.Name != "ctf/web" || resp.Tag != "v1" {
		t.Fatalf("unexpected image resp: %+v", resp)
	}
}

func TestImageServiceTreatsChallengeImageNotFoundAsImageNotFound(t *testing.T) {
	t.Parallel()

	service := NewImageService(&stubChallengeImageRepository{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, challengeports.ErrChallengeImageNotFound
		},
	}, &config.Config{Pagination: config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100}})

	_, err := service.GetImage(context.Background(), 42)
	if err == nil {
		t.Fatal("expected image not found")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrImageNotFound.Code {
		t.Fatalf("expected image not found error, got %v", err)
	}
}

func TestImageServiceListImagesPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeImageContextKey("image-list")
	expectedCtxValue := "ctx-image-list"
	listCalled := false
	repo := &stubChallengeImageRepository{
		listFn: func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-images ctx value %v, got %v", expectedCtxValue, got)
			}
			if name != "web" || status != model.ImageStatusAvailable || offset != 0 || limit != 20 {
				t.Fatalf("unexpected list args: name=%s status=%s offset=%d limit=%d", name, status, offset, limit)
			}
			return []*model.Image{{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable}}, 1, nil
		},
	}
	service := NewImageService(repo, &config.Config{Pagination: config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100}})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.ListImages(ctx, ListImagesInput{Name: "web", Status: model.ImageStatusAvailable})
	if err != nil {
		t.Fatalf("ListImages() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected repository list to be called")
	}
	if resp == nil || resp.Total != 1 || resp.Page != 1 || resp.Size != 20 {
		t.Fatalf("unexpected image list resp: %+v", resp)
	}
}
