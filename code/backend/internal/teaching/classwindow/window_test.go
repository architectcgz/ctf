package classwindow

import (
	"testing"
	"time"
)

func TestParseUsesDefaultSevenDayWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 14, 15, 30, 0, 0, time.UTC)
	window, err := Parse(now, "", "")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if window.FromDate != "2026-05-08" || window.ToDate != "2026-05-14" {
		t.Fatalf("unexpected window dates: %+v", window)
	}
	if window.Days != 7 {
		t.Fatalf("expected default 7 days, got %+v", window)
	}
	if !window.Since.Equal(time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected since: %s", window.Since)
	}
}

func TestParseRejectsSingleBoundary(t *testing.T) {
	t.Parallel()

	if _, err := Parse(time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), "2026-05-01", ""); err == nil {
		t.Fatal("expected single boundary to be rejected")
	}
}

func TestParseRejectsTooLargeRange(t *testing.T) {
	t.Parallel()

	if _, err := Parse(time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), "2026-04-01", "2026-05-14"); err == nil {
		t.Fatal("expected oversized range to be rejected")
	}
}
