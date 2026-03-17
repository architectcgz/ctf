package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ctf-platform/internal/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	tokenLeeway      = time.Second
)

var (
	ErrExpiredToken = errors.New("token expired")
	ErrInvalidToken = errors.New("token invalid")
)

type Claims struct {
	UserID    int64  `json:"uid"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(cfg config.AuthConfig, fallbackIssuer string) (*Manager, error) {
	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	publicKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	issuer := cfg.Issuer
	if issuer == "" {
		issuer = fallbackIssuer
	}

	return &Manager{
		privateKey:      privateKey,
		publicKey:       publicKey,
		issuer:          issuer,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}, nil
}

func (m *Manager) GenerateAccessToken(userID int64, username, role string) (string, *Claims, error) {
	return m.generateToken(userID, username, role, TokenTypeAccess, m.accessTokenTTL)
}

func (m *Manager) GenerateRefreshToken(userID int64, username, role string) (string, *Claims, error) {
	return m.generateToken(userID, username, role, TokenTypeRefresh, m.refreshTokenTTL)
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, ErrInvalidToken
		}
		return m.publicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithLeeway(tokenLeeway))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *Manager) AccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}

func (m *Manager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

func (m *Manager) generateToken(userID int64, username, role, tokenType string, ttl time.Duration) (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newJTI(),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("decode private key: %w", ErrInvalidToken)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	privateKey, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
	if parseErr != nil {
		return nil, fmt.Errorf("parse private key: %w", parseErr)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidToken
	}
	return rsaKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	block, _ := pem.Decode(content)
	if block == nil {
		return nil, ErrInvalidToken
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidToken
	}
	return rsaKey, nil
}

func newJTI() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}
