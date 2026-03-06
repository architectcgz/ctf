package crypto

import (
	"strings"
	"testing"
)

func TestGenerateDynamicFlag(t *testing.T) {
	t.Parallel()

	flag := GenerateDynamicFlag(1, 100, "secret", "nonce123")
	if !strings.HasPrefix(flag, "flag{") || !strings.HasSuffix(flag, "}") {
		t.Fatalf("invalid flag format: %s", flag)
	}
	if len(flag) < 10 {
		t.Fatalf("flag too short: %d", len(flag))
	}

	// 相同输入应生成相同 flag
	flag2 := GenerateDynamicFlag(1, 100, "secret", "nonce123")
	if flag != flag2 {
		t.Fatalf("flags should be identical")
	}

	// 不同 nonce 应生成不同 flag
	flag3 := GenerateDynamicFlag(1, 100, "secret", "nonce456")
	if flag == flag3 {
		t.Fatalf("flags should differ with different nonce")
	}
}

func TestHashStaticFlag(t *testing.T) {
	t.Parallel()

	hash := HashStaticFlag("flag{test}", "salt123")
	if len(hash) != 64 { // SHA256 hex = 64 chars
		t.Fatalf("unexpected hash length: %d", len(hash))
	}

	// 相同输入应生成相同哈希
	hash2 := HashStaticFlag("flag{test}", "salt123")
	if hash != hash2 {
		t.Fatalf("hashes should be identical")
	}

	// 不同 salt 应生成不同哈希
	hash3 := HashStaticFlag("flag{test}", "salt456")
	if hash == hash3 {
		t.Fatalf("hashes should differ with different salt")
	}
}

func TestValidateFlag(t *testing.T) {
	t.Parallel()

	if !ValidateFlag("flag{abc}", "flag{abc}") {
		t.Fatal("identical flags should validate")
	}

	if ValidateFlag("flag{abc}", "flag{xyz}") {
		t.Fatal("different flags should not validate")
	}
}

func TestGenerateSalt(t *testing.T) {
	t.Parallel()

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	if len(salt) == 0 {
		t.Fatal("salt should not be empty")
	}

	// 两次生成应不同
	salt2, _ := GenerateSalt()
	if salt == salt2 {
		t.Fatal("salts should be unique")
	}
}

func TestGenerateNonce(t *testing.T) {
	t.Parallel()

	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error = %v", err)
	}
	if len(nonce) == 0 {
		t.Fatal("nonce should not be empty")
	}

	// 两次生成应不同
	nonce2, _ := GenerateNonce()
	if nonce == nonce2 {
		t.Fatal("nonces should be unique")
	}
}
