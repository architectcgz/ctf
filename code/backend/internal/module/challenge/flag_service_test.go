package challenge

import (
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
)

func TestFlagServiceConfigureStaticFlagAndValidate(t *testing.T) {
	db := setupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:        1,
		Title:     "static-flag",
		Status:    model.ChallengeStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(NewRepository(db), strings.Repeat("s", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureStaticFlag(1, "flag{demo_static}", "flag"); err != nil {
		t.Fatalf("ConfigureStaticFlag() error = %v", err)
	}

	ok, err := service.ValidateFlag(0, 1, "flag{demo_static}", "")
	if err != nil {
		t.Fatalf("ValidateFlag() error = %v", err)
	}
	if !ok {
		t.Fatal("expected static flag validation success")
	}

	cfg, err := service.GetFlagConfig(1)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if cfg.FlagType != model.FlagTypeStatic || cfg.FlagPrefix != "flag" || !cfg.Configured {
		t.Fatalf("unexpected flag config: %+v", cfg)
	}
}

func TestFlagServiceConfigureDynamicFlagAndGenerate(t *testing.T) {
	db := setupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:        2,
		Title:     "dynamic-flag",
		Status:    model.ChallengeStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(NewRepository(db), strings.Repeat("d", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureDynamicFlag(2, "ctf"); err != nil {
		t.Fatalf("ConfigureDynamicFlag() error = %v", err)
	}

	flag, err := service.GenerateDynamicFlag(10, 2, "nonce-1")
	if err != nil {
		t.Fatalf("GenerateDynamicFlag() error = %v", err)
	}
	if !strings.HasPrefix(flag, "ctf{") {
		t.Fatalf("unexpected generated flag: %s", flag)
	}

	ok, err := service.ValidateFlag(10, 2, flag, "nonce-1")
	if err != nil {
		t.Fatalf("ValidateFlag() error = %v", err)
	}
	if !ok {
		t.Fatal("expected dynamic flag validation success")
	}
}
