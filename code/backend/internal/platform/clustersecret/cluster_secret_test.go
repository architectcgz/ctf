package clustersecret

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterContainerFlagSecretRejectsMismatchedActiveFingerprint(t *testing.T) {
	t.Parallel()

	db := newClusterSecretTestDB(t)
	ctx := context.Background()

	if err := RegisterContainerFlagSecret(ctx, db, "active", "cluster-secret-12345678901234567890"); err != nil {
		t.Fatalf("first RegisterContainerFlagSecret() error = %v", err)
	}

	err := RegisterContainerFlagSecret(ctx, db, "active", "different-secret-123456789012345678")
	if err == nil {
		t.Fatal("expected mismatched active secret fingerprint to be rejected")
	}
	if !strings.Contains(err.Error(), "container flag secret fingerprint mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckContainerFlagSecretReturnsMismatchForDifferentKeyID(t *testing.T) {
	t.Parallel()

	db := newClusterSecretTestDB(t)
	ctx := context.Background()

	if err := RegisterContainerFlagSecret(ctx, db, "active", "cluster-secret-12345678901234567890"); err != nil {
		t.Fatalf("RegisterContainerFlagSecret() error = %v", err)
	}

	err := CheckContainerFlagSecret(ctx, db, "next", "cluster-secret-12345678901234567890")
	if err == nil {
		t.Fatal("expected mismatched active key id to be rejected")
	}
	if !strings.Contains(err.Error(), "container flag secret key id mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterContainerFlagSecretKeyringRotatesWhenPreviousActiveKeyIsConfigured(t *testing.T) {
	t.Parallel()

	db := newClusterSecretTestDB(t)
	ctx := context.Background()

	if err := RegisterContainerFlagSecret(ctx, db, "old", "old-cluster-secret-123456789012345678"); err != nil {
		t.Fatalf("RegisterContainerFlagSecret() old error = %v", err)
	}

	err := RegisterContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:   "new",
		ActiveSecret:  "new-cluster-secret-123456789012345678",
		Secrets:       map[string]string{"old": "old-cluster-secret-123456789012345678", "new": "new-cluster-secret-123456789012345678"},
		AllowRotation: true,
	})
	if err != nil {
		t.Fatalf("RegisterContainerFlagSecretKeyring() rotation error = %v", err)
	}
	if err := CheckContainerFlagSecret(ctx, db, "new", "new-cluster-secret-123456789012345678"); err != nil {
		t.Fatalf("CheckContainerFlagSecret() new error = %v", err)
	}
}

func TestRegisterContainerFlagSecretKeyringRejectsRotationWithoutPreviousActiveKey(t *testing.T) {
	t.Parallel()

	db := newClusterSecretTestDB(t)
	ctx := context.Background()

	if err := RegisterContainerFlagSecret(ctx, db, "old", "old-cluster-secret-123456789012345678"); err != nil {
		t.Fatalf("RegisterContainerFlagSecret() old error = %v", err)
	}

	err := RegisterContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:   "new",
		ActiveSecret:  "new-cluster-secret-123456789012345678",
		Secrets:       map[string]string{"new": "new-cluster-secret-123456789012345678"},
		AllowRotation: true,
	})
	if err == nil {
		t.Fatal("expected rotation without previous active key to be rejected")
	}
	if !strings.Contains(err.Error(), "previous active container flag secret is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckContainerFlagSecretKeyringRejectsMissingRequiredPreviousKey(t *testing.T) {
	t.Parallel()

	db := newClusterSecretTestDB(t)
	ctx := context.Background()

	if err := RegisterContainerFlagSecret(ctx, db, "old", "old-cluster-secret-123456789012345678"); err != nil {
		t.Fatalf("RegisterContainerFlagSecret() old error = %v", err)
	}
	if err := RegisterContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:    "new",
		ActiveSecret:   "new-cluster-secret-123456789012345678",
		Secrets:        map[string]string{"old": "old-cluster-secret-123456789012345678", "new": "new-cluster-secret-123456789012345678"},
		RequiredKeyIDs: []string{"old"},
		AllowRotation:  true,
	}); err != nil {
		t.Fatalf("RegisterContainerFlagSecretKeyring() rotation error = %v", err)
	}

	err := CheckContainerFlagSecretKeyring(ctx, db, ContainerFlagSecretKeyring{
		ActiveKeyID:    "new",
		ActiveSecret:   "new-cluster-secret-123456789012345678",
		Secrets:        map[string]string{"new": "new-cluster-secret-123456789012345678"},
		RequiredKeyIDs: []string{"old"},
	})
	if err == nil {
		t.Fatal("expected readiness check to reject a keyring missing the required old key")
	}
	if !strings.Contains(err.Error(), "required container flag secret key old is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newClusterSecretTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&RuntimeClusterSecret{}); err != nil {
		t.Fatalf("migrate cluster secret table: %v", err)
	}
	return db
}
