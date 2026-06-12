package storage

import "testing"

func TestNormalizeKeyRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	unsafeKeys := []string{"", ".", "../escape", "/absolute", "reports/../../escape", "reports\\..\\escape"}
	for _, key := range unsafeKeys {
		key := key
		t.Run(key, func(t *testing.T) {
			t.Parallel()

			if _, err := NormalizeKey(key); err == nil {
				t.Fatalf("NormalizeKey(%q) expected error", key)
			}
		})
	}
}

func TestNormalizeKeyAcceptsRelativeNamespace(t *testing.T) {
	t.Parallel()

	key, err := NormalizeKey(" reports/class/report.pdf ")
	if err != nil {
		t.Fatalf("NormalizeKey() error = %v", err)
	}
	if key != "reports/class/report.pdf" {
		t.Fatalf("key = %q", key)
	}
}
