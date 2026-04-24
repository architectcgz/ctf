package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestServiceListChallengeImportsSortsAndFiltersByActor(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "older-owned",
		FileName:  "older.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "older-owned",
			FileName:   "older.zip",
			Title:      "Older Owned",
			CreatedAt:  time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "static"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "newer-owned",
		FileName:  "newer.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "newer-owned",
			FileName:   "newer.zip",
			Title:      "Newer Owned",
			CreatedAt:  time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
			Category:   "misc",
			Difficulty: "medium",
			Points:     150,
			Flag:       dto.ChallengeImportFlagResp{Type: "dynamic"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "other-user",
		FileName:  "other.zip",
		CreatedBy: 2002,
		CreatedAt: time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "other-user",
			FileName:   "other.zip",
			Title:      "Other User",
			CreatedAt:  time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
			Category:   "crypto",
			Difficulty: "hard",
			Points:     300,
			Flag:       dto.ChallengeImportFlagResp{Type: "static"},
		},
	})

	service := &ChallengeService{}

	previews, err := service.ListChallengeImports(context.Background(), 1001)
	if err != nil {
		t.Fatalf("ListChallengeImports() error = %v", err)
	}

	if len(previews) != 2 {
		t.Fatalf("expected 2 previews, got %d", len(previews))
	}
	if previews[0].ID != "newer-owned" {
		t.Fatalf("expected newest preview first, got %s", previews[0].ID)
	}
	if previews[1].ID != "older-owned" {
		t.Fatalf("expected older preview second, got %s", previews[1].ID)
	}
}

func TestServiceListChallengeImportsReturnsEmptyWhenPreviewRootMissing(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "missing")
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	service := &ChallengeService{}

	previews, err := service.ListChallengeImports(context.Background(), 1001)
	if err != nil {
		t.Fatalf("ListChallengeImports() error = %v", err)
	}
	if len(previews) != 0 {
		t.Fatalf("expected no previews, got %d", len(previews))
	}
}

func TestCommitChallengeImportRestoresSoftDeletedChallenge(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewChallengeService(db, repo, imageRepo, nil, nil, SelfCheckConfig{}, zap.NewNop())

	packageDir := filepath.Join(tempDir, "package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("restored statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	manifest := []byte(`api_version: v1
kind: challenge

meta:
  slug: web-source-audit-double-wrap-01
  title: "Web-01 源码审计：双层伪装"
  category: web
  difficulty: easy
  points: 100

content:
  statement: statement.md

flag:
  type: static
  prefix: flag
  value: flag{web-source-audit-double-wrap-01}
`)
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), manifest, 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}

	deletedAt := time.Date(2026, 4, 9, 20, 37, 35, 0, time.FixedZone("CST", 8*3600))
	legacyChallenge := model.Challenge{
		Title:       "legacy title",
		Description: "legacy description",
		Category:    "web",
		Difficulty:  "easy",
		Points:      50,
		Status:      model.ChallengeStatusPublished,
		PackageSlug: stringPointer("web-source-audit-double-wrap-01"),
		CreatedBy:   int64Pointer(4),
		DeletedAt:   modelDeletedAt(deletedAt),
	}
	if err := db.Unscoped().Create(&legacyChallenge).Error; err != nil {
		t.Fatalf("seed legacy challenge: %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "restore-soft-deleted",
		FileName:  "restore-soft-deleted.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "restore-soft-deleted",
			FileName:   "restore-soft-deleted.zip",
			Slug:       "web-source-audit-double-wrap-01",
			Title:      "Web-01 源码审计：双层伪装",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	if _, err := service.CommitChallengeImport(context.Background(), 4, "restore-soft-deleted"); err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var restored model.Challenge
	if err := db.Unscoped().First(&restored, legacyChallenge.ID).Error; err != nil {
		t.Fatalf("load restored challenge: %v", err)
	}
	if restored.DeletedAt.Valid {
		t.Fatalf("expected soft-deleted challenge to be restored, got deleted_at=%v", restored.DeletedAt.Time)
	}
	if restored.Status != model.ChallengeStatusDraft {
		t.Fatalf("expected restored challenge status to become draft, got %s", restored.Status)
	}
	if restored.Title != "Web-01 源码审计：双层伪装" {
		t.Fatalf("expected restored challenge title to update, got %q", restored.Title)
	}
}

func TestCommitChallengeImportClearsLegacyPublishCheckJobs(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewChallengeService(db, repo, imageRepo, nil, nil, SelfCheckConfig{}, zap.NewNop())

	packageDir := filepath.Join(tempDir, "package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("restored statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	manifest := []byte(`api_version: v1
kind: challenge

meta:
  slug: web-source-audit-double-wrap-01
  title: "Web-01 源码审计：双层伪装"
  category: web
  difficulty: easy
  points: 100

content:
  statement: statement.md

flag:
  type: static
  prefix: flag
  value: flag{web-source-audit-double-wrap-01}
`)
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), manifest, 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}

	challenge := model.Challenge{
		Title:       "legacy title",
		Description: "legacy description",
		Category:    "web",
		Difficulty:  "easy",
		Points:      50,
		Status:      model.ChallengeStatusDraft,
		PackageSlug: stringPointer("web-source-audit-double-wrap-01"),
		CreatedBy:   int64Pointer(4),
	}
	if err := db.Create(&challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	legacyJob := model.ChallengePublishCheckJob{
		ChallengeID:    challenge.ID,
		RequestedBy:    4,
		Status:         model.ChallengePublishCheckStatusFailed,
		RequestSource:  "manual",
		FailureSummary: "单容器拉起失败: Error response from daemon: No such image: registry.example.edu/ctf/web-source-audit-double-wrap-01:20260404",
		CreatedAt:      time.Now().Add(-time.Hour),
		UpdatedAt:      time.Now().Add(-time.Hour),
	}
	if err := db.Create(&legacyJob).Error; err != nil {
		t.Fatalf("seed legacy publish check job: %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "clear-legacy-publish-check-jobs",
		FileName:  "clear-legacy-publish-check-jobs.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "clear-legacy-publish-check-jobs",
			FileName:   "clear-legacy-publish-check-jobs.zip",
			Slug:       "web-source-audit-double-wrap-01",
			Title:      "Web-01 源码审计：双层伪装",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	if _, err := service.CommitChallengeImport(context.Background(), 4, "clear-legacy-publish-check-jobs"); err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var count int64
	if err := db.Model(&model.ChallengePublishCheckJob{}).Where("challenge_id = ?", challenge.ID).Count(&count).Error; err != nil {
		t.Fatalf("count publish check jobs: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected legacy publish check jobs to be cleared, got %d", count)
	}
}

func mustWriteChallengeImportPreviewRecord(t *testing.T, root string, record storedChallengeImportPreview) {
	t.Helper()

	previewDir := filepath.Join(root, record.ID)
	if err := os.MkdirAll(previewDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(previewDir, "preview.json"), content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}

func modelDeletedAt(value time.Time) gorm.DeletedAt {
	return gorm.DeletedAt{Time: value, Valid: true}
}
