package commands

import (
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

func TestFlagServiceConfigureStaticFlagAndValidate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
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

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("s", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureStaticFlag(1, "flag{demo_static}", "flag"); err != nil {
		t.Fatalf("ConfigureStaticFlag() error = %v", err)
	}

	queryService, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("s", 32))
	if err != nil {
		t.Fatalf("NewFlagService(query) error = %v", err)
	}

	ok, err := queryService.ValidateFlag(0, 1, "flag{demo_static}", "")
	if err != nil {
		t.Fatalf("ValidateFlag() error = %v", err)
	}
	if !ok {
		t.Fatal("expected static flag validation success")
	}

	cfg, err := queryService.GetFlagConfig(1)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if cfg.FlagType != model.FlagTypeStatic || cfg.FlagPrefix != "flag" || !cfg.Configured {
		t.Fatalf("unexpected flag config: %+v", cfg)
	}
}

func TestFlagServiceConfigureDynamicFlagAndGenerate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
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

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("d", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureDynamicFlag(2, "ctf"); err != nil {
		t.Fatalf("ConfigureDynamicFlag() error = %v", err)
	}

	queryService, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("d", 32))
	if err != nil {
		t.Fatalf("NewFlagService(query) error = %v", err)
	}

	flag, err := queryService.GenerateDynamicFlag(10, 2, "nonce-1")
	if err != nil {
		t.Fatalf("GenerateDynamicFlag() error = %v", err)
	}
	if !strings.HasPrefix(flag, "ctf{") {
		t.Fatalf("unexpected generated flag: %s", flag)
	}

	ok, err := queryService.ValidateFlag(10, 2, flag, "nonce-1")
	if err != nil {
		t.Fatalf("ValidateFlag() error = %v", err)
	}
	if !ok {
		t.Fatal("expected dynamic flag validation success")
	}
}
