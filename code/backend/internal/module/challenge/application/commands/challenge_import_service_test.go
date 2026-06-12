package commands

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/taxonomy"
	"gorm.io/gorm"
)

func TestServiceListChallengeImportsSortsAndFiltersByActor(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "older-owned",
		FileName:  "older.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "older-owned",
			FileName:   "older.zip",
			Title:      "Older Owned",
			CreatedAt:  time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "newer-owned",
		FileName:  "newer.zip",
		CreatedBy: 1001,
		CreatedAt: time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "newer-owned",
			FileName:   "newer.zip",
			Title:      "Newer Owned",
			CreatedAt:  time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC),
			Category:   "misc",
			Difficulty: "medium",
			Points:     150,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "dynamic"},
		},
	})
	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "other-user",
		FileName:  "other.zip",
		CreatedBy: 2002,
		CreatedAt: time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "other-user",
			FileName:   "other.zip",
			Title:      "Other User",
			CreatedAt:  time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC),
			Category:   "crypto",
			Difficulty: "hard",
			Points:     300,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static"},
		},
	})

	service := newDBBackedChallengeImportService(t, nil, nil)

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

	service := newDBBackedChallengeImportService(t, nil, nil)

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
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

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
	if preview.ImageDelivery.SourceType != challengeentity.ImageSourceTypePlatformBuild {
		t.Fatalf("SourceType = %q, want %q", preview.ImageDelivery.SourceType, challengeentity.ImageSourceTypePlatformBuild)
	}
	if preview.ImageDelivery.TargetImageRef != "127.0.0.1:5000/jeopardy/web-platform-build:v1" {
		t.Fatalf("TargetImageRef = %q", preview.ImageDelivery.TargetImageRef)
	}
	if preview.ImageDelivery.BuildStatus != challengeentity.ImageStatusPending {
		t.Fatalf("BuildStatus = %q, want pending", preview.ImageDelivery.BuildStatus)
	}
}

func TestPreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)

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

	if preview.ImageDelivery.SourceType != challengeentity.ImageSourceTypePlatformBuild {
		t.Fatalf("SourceType = %q, want %q", preview.ImageDelivery.SourceType, challengeentity.ImageSourceTypePlatformBuild)
	}
	if preview.ImageDelivery.TargetImageRef != "" {
		t.Fatalf("expected no target image ref without build service, got %q", preview.ImageDelivery.TargetImageRef)
	}
	if preview.ImageDelivery.BuildStatus != "" {
		t.Fatalf("expected no build status without build service, got %q", preview.ImageDelivery.BuildStatus)
	}
	if len(preview.Warnings) == 0 {
		t.Fatal("expected preview warnings when image build service is unavailable")
	}
	if !challengeImportWarningsContain(preview.Warnings, "当前后端未启用题包镜像构建/校验服务") {
		t.Fatalf("expected service unavailable warning, got %+v", preview.Warnings)
	}
}

func TestCommitChallengeImportCreatesPlatformBuildJob(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

	packageDir := writePlatformBuildChallengePackage(t, tempDir, "web-platform-build")
	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "platform-build",
		FileName:  "platform-build.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "platform-build",
			FileName:   "platform-build.zip",
			Slug:       "web-platform-build",
			Title:      "Web Platform Build",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 4, "platform-build")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}
	if resp.Status != string(challengeentity.ChallengeStatusDraft) {
		t.Fatalf("expected draft challenge, got %q", resp.Status)
	}

	var challenge challengeentity.Challenge
	if err := db.First(&challenge, resp.ID).Error; err != nil {
		t.Fatalf("load challenge: %v", err)
	}
	if challenge.ImageID == nil || *challenge.ImageID <= 0 {
		t.Fatal("expected challenge image id")
	}

	image, err := imageRepo.FindByID(context.Background(), *challenge.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "127.0.0.1:5000/jeopardy/web-platform-build" ||
		image.Tag != "v1" ||
		image.Status != challengeentity.ImageStatusPending ||
		image.SourceType != challengeentity.ImageSourceTypePlatformBuild ||
		image.BuildJobID == nil {
		t.Fatalf("unexpected image: %+v", image)
	}

	job, err := imageRepo.FindImageBuildJobByID(context.Background(), *image.BuildJobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if job.Status != challengeentity.ImageBuildJobStatusPending ||
		job.TargetRef != "127.0.0.1:5000/jeopardy/web-platform-build:v1" {
		t.Fatalf("unexpected build job: %+v", job)
	}
}

func TestCommitChallengeImportDemotesPublishedLegacyChallengeAndPublishesCatalogChangedEvent(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)
	var publishedEvents []platformevents.Event
	service.SetEventBus(&challengeCommandEventBusStub{
		publishFn: func(ctx context.Context, evt platformevents.Event) error {
			publishedEvents = append(publishedEvents, evt)
			return nil
		},
	})

	packageDir := filepath.Join(tempDir, "package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("updated statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	manifest := []byte(`api_version: v1
kind: challenge

meta:
  slug: web-import-refresh
  title: "Legacy Web"
  category: web
  difficulty: easy
  points: 200

content:
  statement: statement.md

flag:
  type: static
  prefix: flag
  value: flag{web-import-refresh}
`)
	if err := os.WriteFile(filepath.Join(packageDir, "challenge.yml"), manifest, 0o644); err != nil {
		t.Fatalf("WriteFile(challenge.yml) error = %v", err)
	}

	legacyChallenge := challengeentity.Challenge{
		Title:       "Legacy Web",
		Description: "legacy",
		Category:    taxonomy.DimensionWeb,
		Difficulty:  challengeentity.ChallengeDifficultyEasy,
		Points:      100,
		Status:      challengeentity.ChallengeStatusPublished,
		CreatedBy:   int64Pointer(4),
	}
	if err := db.Create(&legacyChallenge).Error; err != nil {
		t.Fatalf("seed legacy challenge: %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "legacy-published-refresh",
		FileName:  "legacy-published-refresh.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "legacy-published-refresh",
			FileName:   "legacy-published-refresh.zip",
			Slug:       "web-import-refresh",
			Title:      "Legacy Web",
			Category:   "web",
			Difficulty: "easy",
			Points:     200,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 4, "legacy-published-refresh")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}
	if resp.ID != legacyChallenge.ID {
		t.Fatalf("expected imported challenge to reuse legacy id %d, got %d", legacyChallenge.ID, resp.ID)
	}
	if resp.Status != string(challengeentity.ChallengeStatusDraft) {
		t.Fatalf("expected imported challenge to become draft, got %q", resp.Status)
	}

	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 challenge event, got %+v", publishedEvents)
	}
	if publishedEvents[0].Name != challengecontracts.EventPublishedCatalogChanged {
		t.Fatalf("unexpected event name: %+v", publishedEvents[0])
	}
	payload, ok := publishedEvents[0].Payload.(challengecontracts.PublishedCatalogChangedEvent)
	if !ok {
		t.Fatalf("unexpected event payload type: %T", publishedEvents[0].Payload)
	}
	if payload.ChangeType != challengecontracts.ChallengeCatalogChangeTypeImported ||
		payload.ChallengeID != legacyChallenge.ID ||
		payload.PreviousStatus != challengecontracts.ChallengeStatusPublished ||
		payload.CurrentStatus != challengecontracts.ChallengeStatusDraft ||
		payload.PreviousPoints != 100 ||
		payload.CurrentPoints != 200 {
		t.Fatalf("unexpected event payload: %+v", payload)
	}
}

func TestCommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)

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

	_, err = service.CommitChallengeImport(context.Background(), 4, preview.ID)
	if err == nil {
		t.Fatal("expected commit to fail when image build service is unavailable")
	}
	assertChallengeImportServiceUnavailableError(t, err)
}

func TestCommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)

	packageDir := writeExternalRefChallengePackage(t, tempDir, "web-external-ref", "registry.example.edu/ctf/web-external-ref:v1")
	preview, err := service.PreviewChallengeImport(
		context.Background(),
		4,
		"web-external-ref.zip",
		bytes.NewReader(buildZipArchiveFromDir(t, packageDir)),
	)
	if err != nil {
		t.Fatalf("PreviewChallengeImport() error = %v", err)
	}

	_, err = service.CommitChallengeImport(context.Background(), 4, preview.ID)
	if err == nil {
		t.Fatal("expected commit to fail when external image verification service is unavailable")
	}
	assertChallengeImportServiceUnavailableError(t, err)
}

func TestCommitChallengeImportFromPreviewKeepsPlatformBuildSourceAccessible(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_IMAGE_BUILD_SOURCE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

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

	resp, err := service.CommitChallengeImport(context.Background(), 4, preview.ID)
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var challenge challengeentity.Challenge
	if err := db.First(&challenge, resp.ID).Error; err != nil {
		t.Fatalf("load challenge: %v", err)
	}
	if challenge.ImageID == nil {
		t.Fatal("expected challenge image id")
	}
	image, err := imageRepo.FindByID(context.Background(), *challenge.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.BuildJobID == nil {
		t.Fatal("expected challenge image build job")
	}

	job, err := imageRepo.FindImageBuildJobByID(context.Background(), *image.BuildJobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if _, err := os.Stat(job.ContextPath); err != nil {
		t.Fatalf("expected build context path to exist after commit, got %v", err)
	}
	if _, err := os.Stat(job.DockerfilePath); err != nil {
		t.Fatalf("expected Dockerfile path to exist after commit, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, preview.ID)); !os.IsNotExist(err) {
		t.Fatalf("expected preview dir to be removed after commit, stat err = %v", err)
	}
}

func TestCommitChallengeImportRejectsSoftDeletedDuplicateSlug(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)

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
	legacyChallenge := challengeentity.Challenge{
		Title:       "legacy title",
		Description: "legacy description",
		Category:    "web",
		Difficulty:  "easy",
		Points:      50,
		Status:      challengeentity.ChallengeStatusPublished,
		PackageSlug: stringPointer("web-source-audit-double-wrap-01"),
		CreatedBy:   int64Pointer(4),
		DeletedAt:   modelDeletedAt(deletedAt),
	}
	if err := db.Unscoped().Create(&legacyChallenge).Error; err != nil {
		t.Fatalf("seed legacy challenge: %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "restore-soft-deleted",
		FileName:  "restore-soft-deleted.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "restore-soft-deleted",
			FileName:   "restore-soft-deleted.zip",
			Slug:       "web-source-audit-double-wrap-01",
			Title:      "Web-01 源码审计：双层伪装",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	_, err := service.CommitChallengeImport(context.Background(), 4, "restore-soft-deleted")
	if err == nil {
		t.Fatal("expected soft-deleted duplicate slug import to fail")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}
	if !strings.Contains(appErr.Message, "题目 slug web-source-audit-double-wrap-01 已被已有题目占用") {
		t.Fatalf("unexpected conflict message: %q", appErr.Message)
	}

	var unchanged challengeentity.Challenge
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

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(
		imageRepo,
		ImageBuildConfig{Registry: "127.0.0.1:5000"},
		WithImageBuildDockerBuilder(&fakeDockerImageBuilder{}),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{digest: "sha256:test"}),
	)
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

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

	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "tcp-target",
		FileName:  "tcp-target.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "tcp-target",
			FileName:   "tcp-target.zip",
			Slug:       "pwn-tcp-demo",
			Title:      "Pwn TCP Demo",
			Category:   "pwn",
			Difficulty: "beginner",
			Points:     100,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 4, "tcp-target")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var stored challengeentity.Challenge
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load imported challenge: %v", err)
	}
	if stored.TargetProtocol != challengeentity.ChallengeTargetProtocolTCP {
		t.Fatalf("expected target protocol tcp, got %q", stored.TargetProtocol)
	}
	if stored.TargetPort != 31337 {
		t.Fatalf("expected target port 31337, got %d", stored.TargetPort)
	}
}

func TestCommitChallengeImportRejectsDuplicateSlugWithoutClearingPublishCheckJobs(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengeImportService(t, repo, nil)

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

	challenge := challengeentity.Challenge{
		Title:       "legacy title",
		Description: "legacy description",
		Category:    "web",
		Difficulty:  "easy",
		Points:      50,
		Status:      challengeentity.ChallengeStatusDraft,
		PackageSlug: stringPointer("web-source-audit-double-wrap-01"),
		CreatedBy:   int64Pointer(4),
	}
	if err := db.Create(&challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	legacyJob := challengeentity.ChallengePublishCheckJob{
		ChallengeID:    challenge.ID,
		RequestedBy:    4,
		Status:         challengeentity.ChallengePublishCheckStatusFailed,
		RequestSource:  "manual",
		FailureSummary: "单容器拉起失败: Error response from daemon: No such image: registry.example.edu/ctf/web-source-audit-double-wrap-01:20260404",
		CreatedAt:      time.Now().Add(-time.Hour),
		UpdatedAt:      time.Now().Add(-time.Hour),
	}
	if err := db.Create(&legacyJob).Error; err != nil {
		t.Fatalf("seed legacy publish check job: %v", err)
	}

	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "clear-legacy-publish-check-jobs",
		FileName:  "clear-legacy-publish-check-jobs.zip",
		SourceDir: packageDir,
		CreatedBy: 4,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "clear-legacy-publish-check-jobs",
			FileName:   "clear-legacy-publish-check-jobs.zip",
			Slug:       "web-source-audit-double-wrap-01",
			Title:      "Web-01 源码审计：双层伪装",
			Category:   "web",
			Difficulty: "easy",
			Points:     100,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "static", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	_, err := service.CommitChallengeImport(context.Background(), 4, "clear-legacy-publish-check-jobs")
	if err == nil {
		t.Fatal("expected duplicate slug import to fail")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}

	var count int64
	if err := db.Model(&challengeentity.ChallengePublishCheckJob{}).Where("challenge_id = ?", challenge.ID).Count(&count).Error; err != nil {
		t.Fatalf("count publish check jobs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected legacy publish check jobs to stay untouched after conflict, got %d", count)
	}
}

func TestCommitChallengeImportCreatesTopologyAndPackageRevision(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", tempDir)
	t.Setenv("CHALLENGE_PACKAGE_SOURCE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

	packageDir := writeChallengePackageWithTopology(t, tempDir, "bank-portal")
	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "import-with-topology",
		FileName:  "import-with-topology.zip",
		SourceDir: packageDir,
		CreatedBy: 7,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "import-with-topology",
			FileName:   "import-with-topology.zip",
			Slug:       "bank-portal",
			Title:      "Bank Portal",
			Category:   "web",
			Difficulty: "medium",
			Points:     300,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 7, "import-with-topology")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	var topology challengeentity.ChallengeTopology
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

	spec, err := challengecontracts.DecodeTopologySpec(topology.Spec)
	if err != nil {
		t.Fatalf("DecodeTopologySpec() error = %v", err)
	}
	if len(spec.Nodes) != 2 {
		t.Fatalf("expected 2 topology nodes, got %d", len(spec.Nodes))
	}
	if spec.Nodes[0].ImageID == 0 {
		t.Fatal("expected imported topology node image to resolve to image id")
	}

	var revision challengeentity.ChallengePackageRevision
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
	t.Setenv("CHALLENGE_PACKAGE_SOURCE_DIR", t.TempDir())
	t.Setenv("CHALLENGE_PACKAGE_EXPORT_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := newDBBackedChallengeImportService(t, repo, imageBuildService)

	packageDir := writeChallengePackageWithTopology(t, tempDir, "exportable-bank")
	mustWriteChallengeImportPreviewRecord(t, tempDir, challengeports.ChallengeImportPreviewRecord{
		ID:        "exportable-bank",
		FileName:  "exportable-bank.zip",
		SourceDir: packageDir,
		CreatedBy: 9,
		CreatedAt: time.Now(),
		Preview: challengecontracts.ChallengeImportPreviewResp{
			ID:         "exportable-bank",
			FileName:   "exportable-bank.zip",
			Slug:       "exportable-bank",
			Title:      "Exportable Bank",
			Category:   "web",
			Difficulty: "medium",
			Points:     300,
			Flag:       challengecontracts.ChallengeImportFlagResp{Type: "dynamic", Prefix: "flag"},
			CreatedAt:  time.Now(),
		},
	})

	resp, err := service.CommitChallengeImport(context.Background(), 9, "exportable-bank")
	if err != nil {
		t.Fatalf("CommitChallengeImport() error = %v", err)
	}

	challengeID := resp.ID
	if err := db.Model(&challengeentity.Challenge{}).Where("id = ?", challengeID).Updates(map[string]any{
		"title":  "Exportable Bank v2",
		"points": 450,
	}).Error; err != nil {
		t.Fatalf("update challenge: %v", err)
	}
	var topology challengeentity.ChallengeTopology
	if err := db.Where("challenge_id = ?", challengeID).First(&topology).Error; err != nil {
		t.Fatalf("load topology: %v", err)
	}
	spec, err := challengecontracts.DecodeTopologySpec(topology.Spec)
	if err != nil {
		t.Fatalf("DecodeTopologySpec() error = %v", err)
	}
	spec.Nodes[0].ServicePort = 9090
	topology.Spec, err = challengecontracts.EncodeTopologySpec(spec)
	if err != nil {
		t.Fatalf("EncodeTopologySpec() error = %v", err)
	}
	if err := db.Save(&topology).Error; err != nil {
		t.Fatalf("save topology: %v", err)
	}

	exportService := newDBBackedChallengePackageExportService(repo)
	exportResp, err := exportService.ExportChallengePackage(context.Background(), 9, challengeID)
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

	var refreshed challengeentity.ChallengeTopology
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

func TestGetChallengePackageExportMapsMissingChallengeToChallengeNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengePackageExportService(repo)

	_, err := service.GetChallengePackageExport(context.Background(), 404, nil)
	if err == nil {
		t.Fatal("expected missing challenge error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != challengecontracts.ErrChallengeNotFound.Code {
		t.Fatalf("expected challenge not found app error, got %v", err)
	}
}

func TestGetChallengePackageExportMapsMissingTopologyToNotFound(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newDBBackedChallengePackageExportService(repo)

	challenge := &challengeentity.Challenge{
		Title:      "no-topology",
		Category:   taxonomy.DimensionWeb,
		Difficulty: challengeentity.ChallengeDifficultyEasy,
		Points:     100,
		Status:     challengeentity.ChallengeStatusDraft,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	_, err := service.GetChallengePackageExport(context.Background(), challenge.ID, nil)
	if err == nil {
		t.Fatal("expected missing topology error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrNotFound.Code {
		t.Fatalf("expected not found app error, got %v", err)
	}
}

func mustWriteChallengeImportPreviewRecord(t *testing.T, root string, record challengeports.ChallengeImportPreviewRecord) {
	t.Helper()

	store := challengeinfra.NewChallengeImportPreviewStore(root)
	if err := store.SaveRecord(context.Background(), &record); err != nil {
		t.Fatalf("SaveRecord() error = %v", err)
	}
}

func assertChallengeImportServiceUnavailableError(t *testing.T, err error) {
	t.Helper()

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %v", err)
	}
	if appErr.Code != apperror.ErrServiceUnavailable.Code {
		t.Fatalf("expected service unavailable code, got %+v", appErr)
	}
	if apperror.HTTPStatus(appErr) != apperror.HTTPStatus(apperror.ErrServiceUnavailable) {
		t.Fatalf("expected service unavailable status, got %+v", appErr)
	}
	if !strings.Contains(appErr.Message, "当前后端未启用题包镜像构建/校验服务") {
		t.Fatalf("unexpected service unavailable message: %q", appErr.Message)
	}
}

func challengeImportWarningsContain(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, needle) {
			return true
		}
	}
	return false
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

func writeExternalRefChallengePackage(t *testing.T, root string, slug string, imageRef string) string {
	t.Helper()

	packageDir := filepath.Join(root, slug+"-package")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(packageDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "statement.md"), []byte("external ref statement"), 0o644); err != nil {
		t.Fatalf("WriteFile(statement.md) error = %v", err)
	}
	manifest := `api_version: v1
kind: challenge

meta:
  slug: ` + slug + `
  title: External Ref Challenge
  category: web
  difficulty: easy
  points: 100

content:
  statement: statement.md

flag:
  type: static
  prefix: flag
  value: flag{external-ref}

runtime:
  type: container
  image:
    ref: ` + imageRef + `
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
