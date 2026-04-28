package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestImageRepositoryListExplicitlyFiltersSoftDeletedImages(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("image_repository.go")
	if err != nil {
		t.Fatalf("read image repository: %v", err)
	}

	listStart := strings.Index(string(source), "func (r *ImageRepository) List(")
	if listStart < 0 {
		t.Fatal("ImageRepository.List not found")
	}
	updateStart := strings.Index(string(source)[listStart:], "func (r *ImageRepository) Update(")
	if updateStart < 0 {
		t.Fatal("ImageRepository.Update not found")
	}
	listSource := string(source)[listStart : listStart+updateStart]
	if !strings.Contains(listSource, "deleted_at IS NULL") {
		t.Fatal("ImageRepository.List should explicitly filter deleted_at IS NULL instead of relying on implicit GORM scope")
	}
}
