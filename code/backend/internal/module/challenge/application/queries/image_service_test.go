package queries

import (
	"context"
	"testing"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type stubChallengeImageRepository struct {
	createFn                   func(image *model.Image) error
	createWithContextFn        func(ctx context.Context, image *model.Image) error
	findByIDFn                 func(id int64) (*model.Image, error)
	findByIDWithContextFn      func(ctx context.Context, id int64) (*model.Image, error)
	findByNameTagFn            func(name, tag string) (*model.Image, error)
	findByNameTagWithContextFn func(ctx context.Context, name, tag string) (*model.Image, error)
	listFn                     func(name, status string, offset, limit int) ([]*model.Image, int64, error)
	listWithContextFn          func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error)
	updateFn                   func(image *model.Image) error
	updateWithContextFn        func(ctx context.Context, image *model.Image) error
	deleteFn                   func(id int64) error
	deleteWithContextFn        func(ctx context.Context, id int64) error
}

func (s *stubChallengeImageRepository) Create(image *model.Image) error {
	if s.createFn != nil {
		return s.createFn(image)
	}
	return nil
}

func (s *stubChallengeImageRepository) CreateWithContext(ctx context.Context, image *model.Image) error {
	if s.createWithContextFn != nil {
		return s.createWithContextFn(ctx, image)
	}
	return s.Create(image)
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

func (s *stubChallengeImageRepository) FindByNameTagWithContext(ctx context.Context, name, tag string) (*model.Image, error) {
	if s.findByNameTagWithContextFn != nil {
		return s.findByNameTagWithContextFn(ctx, name, tag)
	}
	return s.FindByNameTag(name, tag)
}

func (s *stubChallengeImageRepository) List(name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(name, status, offset, limit)
	}
	return nil, 0, nil
}

func (s *stubChallengeImageRepository) ListWithContext(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listWithContextFn != nil {
		return s.listWithContextFn(ctx, name, status, offset, limit)
	}
	return s.List(name, status, offset, limit)
}

func (s *stubChallengeImageRepository) Update(image *model.Image) error {
	if s.updateFn != nil {
		return s.updateFn(image)
	}
	return nil
}

func (s *stubChallengeImageRepository) UpdateWithContext(ctx context.Context, image *model.Image) error {
	if s.updateWithContextFn != nil {
		return s.updateWithContextFn(ctx, image)
	}
	return s.Update(image)
}

func (s *stubChallengeImageRepository) Delete(id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *stubChallengeImageRepository) DeleteWithContext(ctx context.Context, id int64) error {
	if s.deleteWithContextFn != nil {
		return s.deleteWithContextFn(ctx, id)
	}
	return s.Delete(id)
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

func TestImageServiceListImagesWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := challengeImageContextKey("image-list")
	expectedCtxValue := "ctx-image-list"
	listCalled := false
	repo := &stubChallengeImageRepository{
		listWithContextFn: func(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
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
	resp, err := service.ListImagesWithContext(ctx, &dto.ImageQuery{Name: "web", Status: model.ImageStatusAvailable})
	if err != nil {
		t.Fatalf("ListImagesWithContext() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected repository list to be called")
	}
	if resp == nil || resp.Total != 1 || resp.Page != 1 || resp.Size != 20 {
		t.Fatalf("unexpected image list resp: %+v", resp)
	}
}
