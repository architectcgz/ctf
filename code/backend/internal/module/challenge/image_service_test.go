package challenge

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type stubImageRuntime struct {
	inspectImageSizeFn func(ctx context.Context, imageRef string) (int64, error)
	removeImageFn      func(ctx context.Context, imageRef string) error
}

func (s *stubImageRuntime) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	if s.inspectImageSizeFn == nil {
		return 0, nil
	}
	return s.inspectImageSizeFn(ctx, imageRef)
}

func (s *stubImageRuntime) RemoveImage(ctx context.Context, imageRef string) error {
	if s.removeImageFn == nil {
		return nil
	}
	return s.removeImageFn(ctx, imageRef)
}

func TestImageServiceDeleteImageReturnsInUseWhenChallengeReferencesImage(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	image := &model.Image{Name: "web", Tag: "latest", Status: model.ImageStatusAvailable}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		Title:   "challenge-1",
		ImageID: image.ID,
		Status:  model.ChallengeStatusDraft,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	service := NewImageService(
		NewImageRepository(db),
		NewRepository(db),
		nil,
		&config.Config{},
		nil,
	)

	err := service.DeleteImage(image.ID)
	if err == nil || err.Error() != errcode.ErrImageInUse.Error() {
		t.Fatalf("expected image in use error, got %v", err)
	}
}

func TestImageServiceDeleteImageRemovesUnreferencedImage(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	image := &model.Image{Name: "web", Tag: "latest", Status: model.ImageStatusAvailable}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := NewImageService(
		NewImageRepository(db),
		NewRepository(db),
		nil,
		&config.Config{},
		nil,
	)

	if err := service.DeleteImage(image.ID); err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}

	var count int64
	if err := db.Model(&model.Image{}).Where("id = ?", image.ID).Count(&count).Error; err != nil {
		t.Fatalf("count image: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected image to be deleted, count = %d", count)
	}
}

func TestImageServiceCreateImageWithContextHonorsCancellation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewImageService(
		NewImageRepository(db),
		NewRepository(db),
		&stubImageRuntime{
			inspectImageSizeFn: func(ctx context.Context, imageRef string) (int64, error) {
				if imageRef != "web:latest" {
					t.Fatalf("unexpected image ref: %s", imageRef)
				}
				<-ctx.Done()
				return 0, ctx.Err()
			},
		},
		&config.Config{},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.CreateImageWithContext(ctx, &dto.CreateImageReq{
		Name: "web",
		Tag:  "latest",
	})
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestImageServiceCloseCancelsAsyncDelete(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	image := &model.Image{Name: "web", Tag: "latest", Status: model.ImageStatusAvailable}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	startedCh := make(chan struct{})
	var removeCalls atomic.Int32
	service := NewImageService(
		NewImageRepository(db),
		NewRepository(db),
		&stubImageRuntime{
			removeImageFn: func(ctx context.Context, imageRef string) error {
				if imageRef != "web:latest" {
					t.Fatalf("unexpected image ref: %s", imageRef)
				}
				removeCalls.Add(1)
				close(startedCh)
				<-ctx.Done()
				return ctx.Err()
			},
		},
		&config.Config{},
		nil,
	)

	if err := service.DeleteImage(image.ID); err != nil {
		t.Fatalf("DeleteImage() error = %v", err)
	}

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected async image removal to start")
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if removeCalls.Load() != 1 {
		t.Fatalf("expected one image removal call, got %d", removeCalls.Load())
	}
}
