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
