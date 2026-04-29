package jobs

import (
	"strings"
	"testing"
)

func TestRedactAndTruncateAWDCheckerText(t *testing.T) {
	input := strings.Repeat("x", 2100) + " flag{round-secret}"

	output, truncated := redactAndTruncateAWDCheckerText(input, 64, "flag{round-secret}")

	if strings.Contains(output, "flag{round-secret}") {
		t.Fatalf("output leaked flag: %q", output)
	}
	if !strings.Contains(output, "[redacted]") {
		t.Fatalf("output was not redacted: %q", output)
	}
	if !truncated {
		t.Fatalf("truncated = false, want true")
	}
	if len(output) > 64+len("...[truncated]") {
		t.Fatalf("output length = %d, want bounded output", len(output))
	}
}
