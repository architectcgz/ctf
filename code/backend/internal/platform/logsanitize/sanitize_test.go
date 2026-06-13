package logsanitize

import (
	"strings"
	"testing"
)

func TestSanitizePasswordAlwaysRedactsValue(t *testing.T) {
	t.Parallel()

	got := SanitizePassword("plain-password")
	if got != RedactedValue {
		t.Fatalf("SanitizePassword() = %q, want %q", got, RedactedValue)
	}
}

func TestSanitizeTokenAlwaysRedactsValue(t *testing.T) {
	t.Parallel()

	token := "session-token-a1b2c3d4e5f6"
	got := SanitizeToken(token)
	if got != RedactedValue {
		t.Fatalf("SanitizeToken() = %q, want %q", got, RedactedValue)
	}
	if strings.Contains(got, token) {
		t.Fatalf("SanitizeToken() leaked token: %q", got)
	}
}

func TestSanitizeSecretAlwaysRedactsValue(t *testing.T) {
	t.Parallel()

	secret := "container-flag-secret-value"
	got := SanitizeSecret(secret)
	if got != RedactedValue {
		t.Fatalf("SanitizeSecret() = %q, want %q", got, RedactedValue)
	}
	if strings.Contains(got, secret) {
		t.Fatalf("SanitizeSecret() leaked secret: %q", got)
	}
}

func TestSanitizeKeyKeepsNamespaceAndShortPrefix(t *testing.T) {
	t.Parallel()

	key := "ctf:auth:session:a1b2c3d4e5f6g7h8"
	got := SanitizeKey(key)

	if !strings.HasPrefix(got, "ctf:auth:session:a1b2") {
		t.Fatalf("SanitizeKey() = %q, want namespace and short prefix", got)
	}
	if strings.Contains(got, "c3d4e5f6g7h8") {
		t.Fatalf("SanitizeKey() leaked full key suffix: %q", got)
	}
	if len(got) >= len(key) {
		t.Fatalf("SanitizeKey() = %q, want shorter than original %q", got, key)
	}
}

func TestSanitizeKeyHandlesEmptyAndShortKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "empty", key: "", want: ""},
		{name: "short", key: "node:web", want: "node:web"},
		{name: "spaces", key: "  session:abcdef1234567890  ", want: "session:abcde..."},
		{name: "short sensitive suffix", key: "session:abcde", want: "session:..."},
		{name: "no namespace", key: "abcdef1234567890", want: "abcdef12..."},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := SanitizeKey(tt.key); got != tt.want {
				t.Fatalf("SanitizeKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
