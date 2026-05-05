package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
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

func TestPreviewChallengeImportReturnsPlatformBuildImageDelivery(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewChallengeService(db, repo, imageRepo, nil, nil, nil, SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)

	packageDir := writePlatformBuildChallengePackage(t, tempDir, "web-platform-build")
	preview, err := service.PreviewChallengeImport(
		context.Background(),
		4,
		"web-platform-build.zip",
		bytes.NewReader(buildZipArchiveFromDir(t, packageDir)),
	)
	if err != nil {
		t.Fatalf("PreviewChallengeImport() error = %v", err)
	}

	if preview.Runtime.ImageRef != "" {
		t.Fatalf("expected no author image ref, got %q", preview.Runtime.ImageRef)
	}
	if preview.ImageDelivery.SourceType != model.ImageSourceTypePlatformBuild {
		t.Fatalf("SourceType = %q, want %q", preview.ImageDelivery.SourceType, model.ImageSourceTypePlatformBuild)
	}
	if preview.ImageDelivery.TargetImageRef != "127.0.0.1:5000/jeopardy/web-platform-build:v1" {
		t.Fatalf("TargetImageRef = %q", preview.ImageDelivery.TargetImageRef)
	}
	if preview.ImageDelivery.BuildStatus != model.ImageStatusPending {
		t.Fatalf("BuildStatus = %q, want pending", preview.ImageDelivery.BuildStatus)
	}
}

func TestCommitChallengeImportCreatesPlatformBuildJob(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewChallengeService(db, repo, imageRepo, nil, nil, nil, SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)

	packageDir := writePlatformBuildChallengePackage(t, tempDir, "web-platform-build")
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "platform-build",
		FileName:  "platform-build.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "platform-build",
			FileName:   "platform-build.zip",
			Slug:       "web-platform-build",
			Title:      "Web Platform Build",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 4, "platform-build")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}
	if resp.Status != model.ChallengeStatusDraft {
		t.Fatalf("expected draft challenge, got %q", resp.Status)
	}

	var challenge model.Challenge
	if err := db.First(&challenge, resp.ID).Error; err != nil {
		t.Fatalf("load challenge: %v", err)
	}
	if challenge.ImageID <= 0 {
		t.Fatal("expected challenge image id")
	}

	image, err := imageRepo.FindByID(context.Background(), challenge.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "127.0.0.1:5000/jeopardy/web-platform-build" ||
		image.Tag != "v1" ||
		image.Status != model.ImageStatusPending ||
		image.SourceType != model.ImageSourceTypePlatformBuild ||
		image.BuildJobID == nil {
		t.Fatalf("unexpected image: %+v", image)
	}

	job, err := imageRepo.FindImageBuildJobByID(context.Background(), *image.BuildJobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if job.Status != model.ImageBuildJobStatusPending ||
		job.TargetRef != "127.0.0.1:5000/jeopardy/web-platform-build:v1" {
		t.Fatalf("unexpected build job: %+v", job)
	}
}

func TestCommitChallengeImportRejectsSoftDeletedDuplicateSlug(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewChallengeService(db, repo, imageRepo, nil, nil, nil, SelfCheckConfig{}, zap.NewNop())

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

	_, err := service.CommitChallengeImport(context.Background(), 4, "restore-soft-deleted")
	if err == nil {
		t.Fatal("expected soft-deleted duplicate slug import to fail")
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}
	if !strings.Contains(appErr.Message, "题目 slug web-source-audit-double-wrap-01 已被已有题目占用") {
		t.Fatalf("unexpected conflict message: %q", appErr.Message)
	}

	var unchanged model.Challenge
	if err := db.Unscoped().First(&unchanged, legacyChallenge.ID).Error; err != nil {
		t.Fatalf("load unchanged challenge: %v", err)
	}
	if !unchanged.DeletedAt.Valid {
		t.Fatalf("expected soft-deleted challenge to stay deleted, got deleted_at=%v", unchanged.DeletedAt.Time)
	}
}

func TestCommitChallengeImportPersistsRuntimeServiceTarget(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(
		imageRepo,
		ImageBuildConfig{Registry: "127.0.0.1:5000"},
		WithImageBuildDockerBuilder(&fakeDockerImageBuilder{}),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{digest: "sha256:test"}),
	)
	service := NewChallengeService(db, repo, imageRepo, nil, nil, nil, SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)

	packageDir := filepath.Join(tempDir, "package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("tcp statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	manifest := []byte(`api_version: v1
kind: challenge

meta:
  slug: pwn-tcp-demo
  title: "Pwn TCP Demo"
  category: pwn
  difficulty: beginner
  points: 100

content:
  statement: statement.md

flag:
  type: static
  prefix: flag
  value: flag{tcp}

runtime:
  type: container
  image:
    ref: 127.0.0.1:5000/ctf/pwn-tcp-demo:v1
  service:
    protocol: tcp
    port: 31337
`)
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), manifest, 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "tcp-target",
		FileName:  "tcp-target.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "tcp-target",
			FileName:   "tcp-target.zip",
			Slug:       "pwn-tcp-demo",
			Title:      "Pwn TCP Demo",
			Category:   "pwn",
			Difficulty: "beginner",
			Points:     100,
			Flag:       dto.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 4, "tcp-target")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var stored model.Challenge
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load imported challenge: %v", err)
	}
	if stored.TargetProtocol != model.ChallengeTargetProtocolTCP {
		t.Fatalf("expected target protocol tcp, got %q", stored.TargetProtocol)
	}
	if stored.TargetPort != 31337 {
		t.Fatalf("expected target port 31337, got %d", stored.TargetPort)
	}
}

func TestCommitChallengeImportRejectsDuplicateSlugWithoutClearingPublishCheckJobs(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	service := NewChallengeService(db, repo, imageRepo, nil, nil, nil, SelfCheckConfig{}, zap.NewNop())

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

	_, err := service.CommitChallengeImport(context.Background(), 4, "clear-legacy-publish-check-jobs")
	if err == nil {
		t.Fatal("expected duplicate slug import to fail")
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}

	var count int64
	if err := db.Model(&model.ChallengePublishCheckJob{}).Where("challenge_id = ?", challenge.ID).Count(&count).Error; err != nil {
		t.Fatalf("count publish check jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected legacy publish check jobs to stay untouched after conflict, got %d", count)
	}
}

func TestCommitChallengeImportCreatesTopologyAndPackageRevision(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())
	t.Setenv("CHALLENGE_PACKAGE_SOURCE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewChallengeService(db, repo, imageRepo, repo, repo, nil, SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)

	packageDir := writeChallengePackageWithTopology(t, tempDir, "bank-portal")
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "import-with-topology",
		FileName:  "import-with-topology.zip",
		SourceDir: packageDir,
		CreatedBy: 7,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "import-with-topology",
			FileName:   "import-with-topology.zip",
			Slug:       "bank-portal",
			Title:      "Bank Portal",
			Category:   "web",
			Difficulty: "medium",
			Points:     300,
			Flag:       dto.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 7, "import-with-topology")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var topology model.ChallengeTopology
	if err := db.Where("challenge_id = ?", resp.ID).First(&topology).Error; err != nil {
		t.Fatalf("load challenge topology: %v", err)
	}
	if topology.EntryNodeKey != "web" {
		t.Fatalf("unexpected entry node key: %q", topology.EntryNodeKey)
	}
	if topology.SourceType != "package_import" {
		t.Fatalf("unexpected topology source type: %q", topology.SourceType)
	}
	if topology.SourcePath != "docker/topology.yml" {
		t.Fatalf("unexpected topology source path: %q", topology.SourcePath)
	}
	if topology.PackageRevisionID == nil || *topology.PackageRevisionID <= 0 {
		t.Fatalf("expected package revision id, got %+v", topology.PackageRevisionID)
	}
	if topology.SyncStatus != "clean" {
		t.Fatalf("expected clean sync status, got %q", topology.SyncStatus)
	}
	if strings.TrimSpace(topology.PackageBaselineSpec) == "" {
		t.Fatal("expected package baseline spec")
	}

	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		t.Fatalf("DecodeTopologySpec() error = %v", err)
	}
	if len(spec.Nodes) != 2 {
		t.Fatalf("expected 2 topology nodes, got %d", len(spec.Nodes))
	}
	if spec.Nodes[0].ImageID == 0 {
		t.Fatal("expected imported topology node image to resolve to image id")
	}

	var revision model.ChallengePackageRevision
	if err := db.First(&revision, *topology.PackageRevisionID).Error; err != nil {
		t.Fatalf("load challenge package revision: %v", err)
	}
	if revision.ChallengeID != resp.ID {
		t.Fatalf("unexpected revision challenge id: %d", revision.ChallengeID)
	}
	if revision.SourceType != "imported" {
		t.Fatalf("unexpected revision source type: %q", revision.SourceType)
	}
	if strings.TrimSpace(revision.SourceDir) == "" {
		t.Fatal("expected revision source dir")
	}
	if _, err := os.Stat(filepath.Join(revision.SourceDir, "docker", "Dockerfile")); err != nil {
		t.Fatalf("expected copied Dockerfile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(revision.SourceDir, "docker", "app.py")); err != nil {
		t.Fatalf("expected copied app.py: %v", err)
	}
}

func TestExportChallengePackageRewritesManifestAndTopology(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())
	t.Setenv("CHALLENGE_PACKAGE_SOURCE_DIR", t.TempDir())
	t.Setenv("CHALLENGE_PACKAGE_EXPORT_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewChallengeService(db, repo, imageRepo, repo, repo, nil, SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)

	packageDir := writeChallengePackageWithTopology(t, tempDir, "exportable-bank")
	mustWriteChallengeImportPreviewRecord(t, tempDir, storedChallengeImportPreview{
		ID:        "exportable-bank",
		FileName:  "exportable-bank.zip",
		SourceDir: packageDir,
		CreatedBy: 9,
		CreatedAt: time.Now(),
		Preview: dto.ChallengeImportPreviewResp{
			ID:         "exportable-bank",
			FileName:   "exportable-bank.zip",
			Slug:       "exportable-bank",
			Title:      "Exportable Bank",
			Category:   "web",
			Difficulty: "medium",
			Points:     300,
			Flag:       dto.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 9, "exportable-bank")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	challengeID := resp.ID
	if err := db.Model(&model.Challenge{}).Where("id = ?", challengeID).Updates(map[string]any{
		"title":  "Exportable Bank v2",
		"points": 450,
	}).Error; err != nil {
		t.Fatalf("update challenge: %v", err)
	}
	var topology model.ChallengeTopology
	if err := db.Where("challenge_id = ?", challengeID).First(&topology).Error; err != nil {
		t.Fatalf("load topology: %v", err)
	}
	spec, err := model.DecodeTopologySpec(topology.Spec)
	if err != nil {
		t.Fatalf("DecodeTopologySpec() error = %v", err)
	}
	spec.Nodes[0].ServicePort = 9090
	topology.Spec, err = model.EncodeTopologySpec(spec)
	if err != nil {
		t.Fatalf("EncodeTopologySpec() error = %v", err)
	}
	if err := db.Save(&topology).Error; err != nil {
		t.Fatalf("save topology: %v", err)
	}

	exportResp, err := service.ExportChallengePackage(context.Background(), 9, challengeID)
	if err != nil {
		t.Fatalf("ExportChallengePackage() error = %v", err)
	}
	if strings.TrimSpace(exportResp.ArchivePath) == "" {
		t.Fatal("expected archive path")
	}
	if _, err := os.Stat(exportResp.ArchivePath); err != nil {
		t.Fatalf("expected export archive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(exportResp.SourceDir, "docker", "Dockerfile")); err != nil {
		t.Fatalf("expected exported Dockerfile: %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(exportResp.SourceDir, "challenge.yml"))
	if err != nil {
		t.Fatalf("read exported challenge.yml: %v", err)
	}
	manifest := string(manifestBytes)
	if !strings.Contains(manifest, "title: Exportable Bank v2") {
		t.Fatalf("expected rewritten title in challenge.yml, got:\n%s", manifest)
	}
	if !strings.Contains(manifest, "points: 450") {
		t.Fatalf("expected rewritten points in challenge.yml, got:\n%s", manifest)
	}

	topologyBytes, err := os.ReadFile(filepath.Join(exportResp.SourceDir, "docker", "topology.yml"))
	if err != nil {
		t.Fatalf("read exported topology.yml: %v", err)
	}
	if !strings.Contains(string(topologyBytes), "service_port: 9090") {
		t.Fatalf("expected rewritten service_port in topology.yml, got:\n%s", string(topologyBytes))
	}

	var refreshed model.ChallengeTopology
	if err := db.Where("challenge_id = ?", challengeID).First(&refreshed).Error; err != nil {
		t.Fatalf("reload topology: %v", err)
	}
	if refreshed.LastExportRevisionID == nil || *refreshed.LastExportRevisionID <= 0 {
		t.Fatalf("expected last export revision id, got %+v", refreshed.LastExportRevisionID)
	}
	if refreshed.PackageRevisionID == nil || *refreshed.PackageRevisionID != *refreshed.LastExportRevisionID {
		t.Fatalf("expected package revision id to move to exported revision, got package=%+v export=%+v", refreshed.PackageRevisionID, refreshed.LastExportRevisionID)
	}
	if refreshed.SyncStatus != "clean" {
		t.Fatalf("expected clean sync status after export, got %q", refreshed.SyncStatus)
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

func buildZipArchiveFromDir(t *testing.T, root string) []byte {
	t.Helper()

	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		writer, err := archive.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		t.Fatalf("walk package dir: %v", err)
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close zip archive: %v", err)
	}
	return buf.Bytes()
}

func writePlatformBuildChallengePackage(t *testing.T, root string, slug string) string {
	t.Helper()

	packageDir := filepath.Join(root, slug+"-package")
	if err := os.MkdirAll(filepath.Join(packageDir, "docker"), 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir/docker) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("platform build statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "docker", "Dockerfile"), []byte("FROM nginx:1.27-alpine"), 0o644); err != nil {
		t.Fatalf("WriteFile(Dockerfile) error = %v", err)
	}
	manifest := `api_version: v1
kind: challenge

meta:
  slug: ` + slug + `
  title: Web Platform Build
  category: web
  difficulty: easy
  points: 100

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: flag

runtime:
  type: container
  image:
    tag: v1
  service:
    protocol: http
    port: 8080
`
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}
	return packageDir
}

func int64Pointer(value int64) *int64 {
	return &value
}

func modelDeletedAt(value time.Time) gorm.DeletedAt {
	return gorm.DeletedAt{Time: value, Valid: true}
}

func writeChallengePackageWithTopology(t *testing.T, root string, slug string) string {
	t.Helper()

	packageDir := filepath.Join(root, slug+"-package")
	if err := os.MkdirAll(filepath.Join(packageDir, "docker"), 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir/docker) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("bank portal statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "docker", "Dockerfile"), []byte("FROM python:3.12-alpine"), 0o644); err != nil {
		t.Fatalf("WriteFile(Dockerfile) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "docker", "app.py"), []byte("print('bank')"), 0o644); err != nil {
		t.Fatalf("WriteFile(app.py) error = %v", err)
	}
	topology := `api_version: v1
kind: topology
entry_node_key: web
networks:
  - key: public
    name: Public
  - key: internal
    name: Internal
    internal: true
nodes:
  - key: web
    name: Web
    tier: public
    image:
      ref: ctf/` + slug + `:web
      dockerfile: docker/Dockerfile
      context: .
    service_port: 8080
    inject_flag: true
    network_keys: [public, internal]
    env:
      APP_ENV: prod
  - key: db
    name: Database
    tier: internal
    image:
      ref: mysql:8.0
    service_port: 3306
    network_keys: [internal]
links:
  - from_node_key: web
    to_node_key: db
policies:
  - source_node_key: web
    target_node_key: db
    action: allow
    protocol: tcp
    ports: [3306]
`
	if err := os.WriteFile(filepath.Join(packageDir, "docker", "topology.yml"), []byte(topology), 0o644); err != nil {
		t.Fatalf("WriteFile(topology.yml) error = %v", err)
	}
	manifest := `api_version: v1
kind: challenge

meta:
  slug: ` + slug + `
  title: ` + strings.ReplaceAll(slug, "-", " ") + `
  category: web
  difficulty: medium
  points: 300

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: flag

runtime:
  type: container
  image:
    ref: ctf/` + slug + `:web

extensions:
  topology:
    enabled: true
    source: docker/topology.yml
`
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}
	return packageDir
}
