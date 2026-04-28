package domain

import (
	"testing"

	"ctf-platform/internal/model"
)

func TestImageRespFromModelIncludesFormattedSize(t *testing.T) {
	t.Parallel()

	resp := ImageRespFromModel(&model.Image{Size: 268435456})
	if resp.SizeFormatted != "256 MB" {
		t.Fatalf("expected formatted image size, got %q", resp.SizeFormatted)
	}
}
