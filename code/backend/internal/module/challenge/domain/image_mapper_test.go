package domain

import (
	"testing"

	challengeentity "ctf-platform/internal/module/challenge/entity"
)

func TestImageRespFromModelIncludesFormattedSize(t *testing.T) {
	t.Parallel()

	resp := ImageRespFromModel(&challengeentity.Image{Size: 268435456})
	if resp.SizeFormatted != "256 MB" {
		t.Fatalf("expected formatted image size, got %q", resp.SizeFormatted)
	}
}
