package challenge

import (
	"testing"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

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
