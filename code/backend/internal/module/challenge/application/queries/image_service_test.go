package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

type stubChallengeImageRepository struct {
	findByIDFn            func(id int64) (*model.Image, error)
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn       func(name, tag string) (*model.Image, error)
	listFn                func(name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn              func(image *model.Image) error
	deleteFn              func(id int64) error
	createFn              func(image *model.Image) error
}

func (s *stubChallengeImageRepository) Create(image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(image)
	}
	return nil
}

func (s *stubChallengeImageRepository) FindByID(id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(id)
	}
	return nil, nil
}

func (s *stubChallengeImageRepository) FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *stubChallengeImageRepository) FindByNameTag(name, tag string) (*model.Image, error) {
	if s.findByNameTagFn != nil {
		return s.findByNameTagFn(name, tag)
	}
	return nil, nil
}

func (s *stubChallengeImageRepository) List(name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *stubChallengeImageRepository) Update(image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(image)
	}
	return nil
}

func (s *stubChallengeImageRepository) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

type challengeImageContextKey string

func TestImageServiceGetImageWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeImageContextKey("image-query")
	expectedCtxValue := "ctx-image-query"
	findCalled := false
	repo := &stubChallengeImageRepository{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
			findCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-by-id ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable}, nil
		},
	}
	service := NewImageService(repo, &config.Config{Pagination: config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100}})

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.GetImageWithContext(ctx, 42)
	if err != nil {
		t.Fatalf("GetImageWithContext() error = %v", err)
	}
	if !findCalled {
		t.Fatal("expected repository find to be called")
	}
	if resp == nil || resp.ID != 42 || resp.Name != "ctf/web" || resp.Tag != "v1" {
		t.Fatalf("unexpected image resp: %+v", resp)
	}
}
