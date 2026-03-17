package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ctf-platform/internal/config"
)

func TestManagerParseTokenAllowsSmallClockSkew(t *testing.T) {
	privatePath, publicPath := writeTestKeyPair(t)
	manager, err := NewManager(config.AuthConfig{
		PrivateKeyPath:  privatePath,
		PublicKeyPath:   publicPath,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}, "ctf-platform-test")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	now := time.Now()
	claims := &Claims{
		UserID:    1,
		Username:  "admin",
		Role:      "admin",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "jti-skew",
			Issuer:    "ctf-platform-test",
			Subject:   "1",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(500 * time.Millisecond)),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(manager.privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	parsed, err := manager.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if parsed.ID != claims.ID {
		t.Fatalf("expected jti %q, got %q", claims.ID, parsed.ID)
	}
}

func TestManagerParseTokenRejectsTokenBeyondClockSkew(t *testing.T) {
	privatePath, publicPath := writeTestKeyPair(t)
	manager, err := NewManager(config.AuthConfig{
		PrivateKeyPath:  privatePath,
		PublicKeyPath:   publicPath,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
	}, "ctf-platform-test")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	now := time.Now()
	claims := &Claims{
		UserID:    1,
		Username:  "admin",
		Role:      "admin",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "jti-future",
			Issuer:    "ctf-platform-test",
			Subject:   "1",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(3 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(manager.privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = manager.ParseToken(tokenString)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func writeTestKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})

	keyDir := t.TempDir()
	privatePath := filepath.Join(keyDir, "private.pem")
	publicPath := filepath.Join(keyDir, "public.pem")
	if err := os.WriteFile(privatePath, privatePEM, 0o600); err != nil {
		t.Fatalf("WriteFile(private) error = %v", err)
	}
	if err := os.WriteFile(publicPath, publicPEM, 0o644); err != nil {
		t.Fatalf("WriteFile(public) error = %v", err)
	}
	return privatePath, publicPath
}
