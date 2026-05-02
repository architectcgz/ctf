package mapperhelper

import "testing"

func TestNormalizeOptionalString(t *testing.T) {
	if got := NormalizeOptionalString(""); got != nil {
		t.Fatalf("expected nil for empty string, got %v", *got)
	}

	got := NormalizeOptionalString("value")
	if got == nil || *got != "value" {
		t.Fatalf("expected pointer to value, got %#v", got)
	}
}

func TestNormalizeOptionalTrimmedString(t *testing.T) {
	if got := NormalizeOptionalTrimmedString("   "); got != nil {
		t.Fatalf("expected nil for blank string, got %v", *got)
	}

	got := NormalizeOptionalTrimmedString("  value  ")
	if got == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if *got != "value" {
		t.Fatalf("expected trimmed value, got %q", *got)
	}
}
