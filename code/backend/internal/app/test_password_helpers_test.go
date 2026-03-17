package app

import (
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"ctf-platform/internal/model"
)

var testPasswordHashCache sync.Map

func setTestPassword(t *testing.T, user *model.User, password string) {
	t.Helper()

	if user == nil {
		t.Fatal("setTestPassword() got nil user")
	}

	hash, err := testPasswordHash(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user.PasswordHash = hash
}

func testPasswordHash(password string) (string, error) {
	if cached, ok := testPasswordHashCache.Load(password); ok {
		return cached.(string), nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	hashValue := string(hash)
	testPasswordHashCache.Store(password, hashValue)
	return hashValue, nil
}
