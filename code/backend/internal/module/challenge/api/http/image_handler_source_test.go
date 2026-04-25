package http

import (
	"os"
	"strings"
	"testing"
)

func TestImageHandlerInvalidIDMessageIsDeclaredOnce(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("image_handler.go")
	if err != nil {
		t.Fatalf("read image handler: %v", err)
	}
	if count := strings.Count(string(source), `"无效的镜像 ID"`); count != 1 {
		t.Fatalf("expected invalid image ID message literal to be declared once, got %d", count)
	}
}
