package commands

import (
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
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

func TestFlagServiceConfigureDynamicFlagRejectsSharedChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:              22,
		Title:           "shared-dynamic-flag",
		Status:          model.ChallengeStatusDraft,
		InstanceSharing: model.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("d", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	err = service.ConfigureDynamicFlag(22, "flag")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared dynamic flag, got %v", err)
	}
}

func TestFlagServiceConfigureSharedProofFlagAllowsSharedChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:              23,
		Title:           "shared-proof-flag",
		Status:          model.ChallengeStatusDraft,
		InstanceSharing: model.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("p", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureSharedProofFlag(23); err != nil {
		t.Fatalf("ConfigureSharedProofFlag() error = %v", err)
	}

	queryService, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("p", 32))
	if err != nil {
		t.Fatalf("NewFlagService(query) error = %v", err)
	}

	cfg, err := queryService.GetFlagConfig(23)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if cfg.FlagType != model.FlagTypeSharedProof || !cfg.Configured {
		t.Fatalf("unexpected shared proof flag config: %+v", cfg)
	}
}

func TestFlagServiceConfigureSharedProofFlagRejectsNonSharedChallenge(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:              24,
		Title:           "non-shared-proof-flag",
		Status:          model.ChallengeStatusDraft,
		InstanceSharing: model.InstanceSharingPerUser,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("p", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	err = service.ConfigureSharedProofFlag(24)
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for non-shared shared_proof flag, got %v", err)
	}
}

func TestFlagServiceConfigureRegexFlagAndValidate(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:        3,
		Title:     "regex-flag",
		Status:    model.ChallengeStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("r", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureRegexFlag(3, `^flag\{user-[0-9]{3}\}$`, "flag"); err != nil {
		t.Fatalf("ConfigureRegexFlag() error = %v", err)
	}

	queryService, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("r", 32))
	if err != nil {
		t.Fatalf("NewFlagService(query) error = %v", err)
	}

	ok, err := queryService.ValidateFlag(0, 3, "flag{user-123}", "")
	if err != nil {
		t.Fatalf("ValidateFlag() error = %v", err)
	}
	if !ok {
		t.Fatal("expected regex flag validation success")
	}

	cfg, err := queryService.GetFlagConfig(3)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if cfg.FlagType != model.FlagTypeRegex || cfg.FlagRegex != `^flag\{user-[0-9]{3}\}$` || !cfg.Configured {
		t.Fatalf("unexpected regex flag config: %+v", cfg)
	}
}

func TestFlagServiceConfigureManualReviewFlag(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Challenge{
		ID:        4,
		Title:     "manual-review",
		Status:    model.ChallengeStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	service, err := NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("m", 32))
	if err != nil {
		t.Fatalf("NewFlagService() error = %v", err)
	}

	if err := service.ConfigureManualReviewFlag(4); err != nil {
		t.Fatalf("ConfigureManualReviewFlag() error = %v", err)
	}

	cfg, err := challengeqry.NewFlagService(challengeinfra.NewRepository(db), strings.Repeat("m", 32))
	if err != nil {
		t.Fatalf("NewFlagService(query) error = %v", err)
	}

	flagCfg, err := cfg.GetFlagConfig(4)
	if err != nil {
		t.Fatalf("GetFlagConfig() error = %v", err)
	}
	if flagCfg.FlagType != model.FlagTypeManualReview || !flagCfg.Configured {
		t.Fatalf("unexpected manual review flag config: %+v", flagCfg)
	}
}
