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

	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
	"ctf-platform/pkg/errcode"
	"gorm.io/gorm"
)

func newAWDChallengeImportServiceForTest(db *gorm.DB, repo *challengeinfra.Repository) *AWDChallengeImportService {
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(
		imageRepo,
		ImageBuildConfig{Registry: "127.0.0.1:5000"},
		WithImageBuildDockerBuilder(&fakeDockerImageBuilder{}),
		WithImageBuildRegistryVerifier(fakeRegistryVerifier{digest: "sha256:test"}),
	)
	return NewAWDChallengeImportService(db, repo, imageBuildService)
}

func TestAWDChallengeImportFlowPreviewAndCommit(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newAWDChallengeImportServiceForTest(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-bank-portal-01.zip",
		bytes.NewReader(buildAWDChallengeImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}

	if preview.ID == "" || preview.Slug != "awd-bank-portal-01" {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if preview.ServiceType != "web_http" || preview.CheckerType != "http_standard" {
		t.Fatalf("unexpected preview awd fields: %+v", preview)
	}
	if preview.FlagMode != "dynamic_team" || preview.DefenseEntryMode != "http" {
		t.Fatalf("unexpected preview imported strategy: %+v", preview)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}

	if committed.ID == 0 || committed.Slug != "awd-bank-portal-01" {
		t.Fatalf("unexpected committed challenge: %+v", committed)
	}
	if committed.Status != "published" {
		t.Fatalf("expected published imported challenge, got %+v", committed)
	}
	if committed.RuntimeConfig["image_id"] == nil {
		t.Fatalf("expected runtime_config.image_id in committed challenge, got %+v", committed.RuntimeConfig)
	}

	stored, err := repo.FindAWDChallengeByID(context.Background(), committed.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if string(stored.Status) != "published" {
		t.Fatalf("unexpected stored status: %+v", stored)
	}

	var accessConfig map[string]any
	if err := json.Unmarshal([]byte(stored.AccessConfig), &accessConfig); err != nil {
		t.Fatalf("unmarshal access_config: %v", err)
	}
	if accessConfig["service_port"] != float64(8080) {
		t.Fatalf("unexpected stored access_config: %+v", accessConfig)
	}

	var runtimeConfig map[string]any
	if err := json.Unmarshal([]byte(stored.RuntimeConfig), &runtimeConfig); err != nil {
		t.Fatalf("unmarshal runtime_config: %v", err)
	}
	if runtimeConfig["image_ref"] != "registry.example.edu/ctf/awd-bank-portal:v1" {
		t.Fatalf("unexpected stored runtime_config: %+v", runtimeConfig)
	}
}

func TestAWDChallengeImportPreviewReturnsPlatformBuildImageDelivery(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewAWDChallengeImportService(db, repo, imageBuildService)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-platform-build.zip",
		bytes.NewReader(buildAWDPlatformBuildImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}

	if preview.ImageDelivery.SourceType != model.ImageSourceTypePlatformBuild {
		t.Fatalf("SourceType = %q, want %q", preview.ImageDelivery.SourceType, model.ImageSourceTypePlatformBuild)
	}
	if preview.ImageDelivery.TargetImageRef != "127.0.0.1:5000/awd/awd-platform-build:c1" {
		t.Fatalf("TargetImageRef = %q", preview.ImageDelivery.TargetImageRef)
	}
	if preview.ImageDelivery.BuildStatus != model.ImageStatusPending {
		t.Fatalf("BuildStatus = %q, want pending", preview.ImageDelivery.BuildStatus)
	}
	if imageRef, _ := preview.RuntimeConfig["image_ref"].(string); imageRef != "" {
		t.Fatalf("expected no author image_ref in preview runtime_config, got %q", imageRef)
	}
}

func TestAWDChallengeImportCommitCreatesPlatformBuildJob(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := NewImageBuildService(imageRepo, ImageBuildConfig{Registry: "127.0.0.1:5000"})
	service := NewAWDChallengeImportService(db, repo, imageBuildService)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-platform-build.zip",
		bytes.NewReader(buildAWDPlatformBuildImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}
	if committed.ReadinessStatus != string(model.AWDReadinessStatusPending) {
		t.Fatalf("ReadinessStatus = %q, want pending", committed.ReadinessStatus)
	}
	if committed.RuntimeConfig["image_ref"] != "127.0.0.1:5000/awd/awd-platform-build:c1" {
		t.Fatalf("unexpected runtime_config.image_ref: %+v", committed.RuntimeConfig)
	}
	imageID := readInt64FromAnyForAWDImportTest(committed.RuntimeConfig["image_id"])
	if imageID <= 0 {
		t.Fatalf("expected runtime_config.image_id, got %+v", committed.RuntimeConfig)
	}

	image, err := imageRepo.FindByID(context.Background(), imageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Status != model.ImageStatusPending ||
		image.SourceType != model.ImageSourceTypePlatformBuild ||
		image.BuildJobID == nil {
		t.Fatalf("unexpected platform image: %+v", image)
	}

	job, err := imageRepo.FindImageBuildJobByID(context.Background(), *image.BuildJobID)
	if err != nil {
		t.Fatalf("FindImageBuildJobByID() error = %v", err)
	}
	if job.Status != model.ImageBuildJobStatusPending ||
		job.TargetRef != "127.0.0.1:5000/awd/awd-platform-build:c1" {
		t.Fatalf("unexpected build job: %+v", job)
	}
}

func TestAWDChallengeImportRejectsDuplicateSlug(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newAWDChallengeImportServiceForTest(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)

	firstPreview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-bank-portal-01.zip",
		bytes.NewReader(buildAWDChallengeImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport(first) error = %v", err)
	}
	firstCommitted, err := service.CommitImport(context.Background(), 2001, firstPreview.ID)
	if err != nil {
		t.Fatalf("CommitImport(first) error = %v", err)
	}

	secondPreview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-bank-portal-01-v2.zip",
		bytes.NewReader(buildAWDChallengeImportArchiveWithMeta(t, "awd-bank-portal-01", "Bank Portal AWD v2", "registry.example.edu/ctf/awd-bank-portal:v2")),
	)
	if err != nil {
		t.Fatalf("PreviewImport(second) error = %v", err)
	}
	_, err = service.CommitImport(context.Background(), 2001, secondPreview.ID)
	if err == nil {
		t.Fatal("expected duplicate awd slug commit to fail")
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrConflict.Code {
		t.Fatalf("expected conflict app error, got %v", err)
	}
	if !strings.Contains(appErr.Message, "AWD 题目 slug awd-bank-portal-01 已被已有题目占用") {
		t.Fatalf("unexpected conflict message: %q", appErr.Message)
	}

	stored, err := repo.FindAWDChallengeByID(context.Background(), firstCommitted.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.Name != "Bank Portal AWD" {
		t.Fatalf("expected original awd challenge name to stay unchanged, got %q", stored.Name)
	}
}

func TestAWDChallengeImportStoresScriptCheckerArtifactPrivately(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newAWDChallengeImportServiceForTest(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	artifactDir := filepath.Join(t.TempDir(), "checker-artifacts")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"script-checker.zip",
		bytes.NewReader(buildAWDScriptCheckerImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}
	if preview.CheckerType != "script_checker" {
		t.Fatalf("CheckerType = %q, want script_checker", preview.CheckerType)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}
	stored, err := repo.FindAWDChallengeByID(context.Background(), committed.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	var checkerConfig map[string]any
	if err := json.Unmarshal([]byte(stored.CheckerConfig), &checkerConfig); err != nil {
		t.Fatalf("unmarshal checker_config: %v", err)
	}
	artifact, ok := checkerConfig["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("expected private artifact metadata in checker_config: %+v", checkerConfig)
	}
	if artifact["entry"] != "docker/check/check.py" {
		t.Fatalf("unexpected artifact entry: %+v", artifact)
	}
	storagePath, _ := artifact["storage_path"].(string)
	if storagePath == "" {
		t.Fatalf("expected artifact storage_path: %+v", artifact)
	}
	if !strings.Contains(storagePath, artifactDir) {
		t.Fatalf("unexpected artifact storage path: %s", storagePath)
	}
	if _, err := os.Stat(storagePath); err != nil {
		t.Fatalf("expected stored checker artifact file: %v", err)
	}
}

func TestAWDChallengeImportStoresScriptCheckerArtifactFiles(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newAWDChallengeImportServiceForTest(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	artifactDir := filepath.Join(t.TempDir(), "checker-artifacts")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"script-checker-files.zip",
		bytes.NewReader(buildAWDMultiFileScriptCheckerImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}
	stored, err := repo.FindAWDChallengeByID(context.Background(), committed.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	var checkerConfig map[string]any
	if err := json.Unmarshal([]byte(stored.CheckerConfig), &checkerConfig); err != nil {
		t.Fatalf("unmarshal checker_config: %v", err)
	}
	artifact, ok := checkerConfig["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("expected private artifact metadata in checker_config: %+v", checkerConfig)
	}
	files, ok := artifact["files"].([]any)
	if !ok || len(files) != 2 {
		t.Fatalf("expected two artifact files: %+v", artifact)
	}
	for _, item := range files {
		file, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("unexpected artifact file item: %#v", item)
		}
		storagePath, _ := file["storage_path"].(string)
		if storagePath == "" || !strings.Contains(storagePath, artifactDir) {
			t.Fatalf("unexpected artifact file storage path: %+v", file)
		}
		if _, err := os.Stat(storagePath); err != nil {
			t.Fatalf("expected stored checker artifact file: %v", err)
		}
	}
}

func TestAWDChallengeImportKeepsTCPStandardCheckerConfig(t *testing.T) {
	db := testsupport.SetupTestDB(t)
	repo := challengeinfra.NewRepository(db)
	service := newAWDChallengeImportServiceForTest(db, repo)

	previewDir := filepath.Join(t.TempDir(), "awd-imports")
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", previewDir)

	preview, err := service.PreviewImport(
		context.Background(),
		2001,
		"awd-tcp-length-gate.zip",
		bytes.NewReader(buildAWDTCPCheckerImportArchive(t)),
	)
	if err != nil {
		t.Fatalf("PreviewImport() error = %v", err)
	}
	if preview.ServiceType != "binary_tcp" || preview.CheckerType != "tcp_standard" {
		t.Fatalf("unexpected preview awd fields: %+v", preview)
	}

	committed, err := service.CommitImport(context.Background(), 2001, preview.ID)
	if err != nil {
		t.Fatalf("CommitImport() error = %v", err)
	}
	stored, err := repo.FindAWDChallengeByID(context.Background(), committed.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if stored.ServiceType != "binary_tcp" || stored.CheckerType != "tcp_standard" {
		t.Fatalf("unexpected stored awd fields: %+v", stored)
	}

	var checkerConfig map[string]any
	if err := json.Unmarshal([]byte(stored.CheckerConfig), &checkerConfig); err != nil {
		t.Fatalf("unmarshal checker_config: %v", err)
	}
	steps, ok := checkerConfig["steps"].([]any)
	if !ok || len(steps) != 3 {
		t.Fatalf("unexpected tcp checker steps: %+v", checkerConfig)
	}
	if checkerConfig["timeout_ms"] != float64(3000) {
		t.Fatalf("unexpected tcp checker timeout: %+v", checkerConfig)
	}
}

func readAWDCheckerArtifactDigestForTest(t *testing.T, checkerConfigRaw string) string {
	t.Helper()
	var checkerConfig map[string]any
	if err := json.Unmarshal([]byte(checkerConfigRaw), &checkerConfig); err != nil {
		t.Fatalf("unmarshal checker_config: %v", err)
	}
	artifact, ok := checkerConfig["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("expected artifact metadata: %+v", checkerConfig)
	}
	digest, _ := artifact["digest"].(string)
	if digest == "" {
		t.Fatalf("expected artifact digest: %+v", artifact)
	}
	return digest
}

func buildAWDChallengeImportArchive(t *testing.T) []byte {
	t.Helper()

	return buildAWDChallengeImportArchiveWithMeta(
		t,
		"awd-bank-portal-01",
		"Bank Portal AWD",
		"registry.example.edu/ctf/awd-bank-portal:v1",
	)
}

func buildAWDChallengeImportArchiveWithMeta(t *testing.T, slug, title, imageRef string) []byte {
	t.Helper()

	files := map[string]string{
		slug + "/challenge.yml": `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: ` + slug + `
  title: ` + title + `
  category: web
  difficulty: hard
  points: 500

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: ` + imageRef + `

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    version: v2026.04
    checker:
      type: http_standard
      config:
        put_flag:
          method: PUT
          path: /api/flag
          expected_status: 200
          body_template: "{{FLAG}}"
        get_flag:
          method: GET
          path: /api/flag
          expected_status: 200
          expected_substring: "{{FLAG}}"
        havoc:
          method: GET
          path: /healthz
          expected_status: 200
    flag_policy:
      mode: dynamic_team
      config:
        flag_prefix: awd
        rotate_interval_sec: 120
    defense_entry:
      mode: http
    access_config:
      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
      exposed_ports:
        - port: 8080
          protocol: tcp
          purpose: http
    runtime_config:
      instance_sharing: per_team
      service_port: 8080
      defense_scope:
        editable_paths:
          - docker/challenge_app.py
        protected_paths:
          - docker/app.py
          - docker/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`,
		slug + "/statement.md":            "银行门户存在越权修改 flag 的逻辑。",
		slug + "/docker/app.py":           "print('entry')\n",
		slug + "/docker/ctf_runtime.py":   "print('runtime')\n",
		slug + "/docker/challenge_app.py": "print('challenge')\n",
		slug + "/docker/check/check.py":   "print('check')\n",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func buildAWDPlatformBuildImportArchive(t *testing.T) []byte {
	t.Helper()

	files := map[string]string{
		"awd-platform-build/challenge.yml": `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: awd-platform-build
  title: AWD Platform Build
  category: web
  difficulty: hard
  points: 500

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    tag: c1

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    version: v2026.05
    checker:
      type: http_standard
      config:
        put_flag:
          method: PUT
          path: /api/flag
          expected_status: 200
          body_template: "{{FLAG}}"
        get_flag:
          method: GET
          path: /api/flag
          expected_status: 200
          expected_substring: "{{FLAG}}"
        havoc:
          method: GET
          path: /healthz
          expected_status: 200
    flag_policy:
      mode: dynamic_team
    defense_entry:
      mode: http
    access_config:
      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
    runtime_config:
      instance_sharing: per_team
      service_port: 8080
      defense_scope:
        editable_paths:
          - docker/challenge_app.py
        protected_paths:
          - docker/app.py
          - docker/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`,
		"awd-platform-build/statement.md":            "平台构建 AWD 服务。",
		"awd-platform-build/docker/Dockerfile":       "FROM python:3.12-alpine\nCOPY . /app\n",
		"awd-platform-build/docker/app.py":           "print('entry')\n",
		"awd-platform-build/docker/ctf_runtime.py":   "print('runtime')\n",
		"awd-platform-build/docker/challenge_app.py": "print('challenge')\n",
		"awd-platform-build/docker/check/check.py":   "print('check')\n",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func buildAWDTCPCheckerImportArchive(t *testing.T) []byte {
	t.Helper()

	files := map[string]string{
		"awd-tcp-length-gate/challenge.yml": `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: awd-tcp-length-gate
  title: TCP Length Gate
  category: pwn
  difficulty: medium

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: registry.example.edu/ctf/awd-tcp-length-gate:v1

extensions:
  awd:
    service_type: binary_tcp
    deployment_mode: single_container
    checker:
      type: tcp_standard
      config:
        timeout_ms: 3000
        steps:
          - send: "PING\n"
            expect_contains: PONG
          - send_template: "SET_FLAG {{FLAG}}\n"
            expect_contains: OK
          - send: "GET_FLAG\n"
            expect_contains: "{{FLAG}}"
    flag_policy:
      mode: dynamic_team
    defense_entry:
      mode: tcp
    access_config:
      public_base_url: tcp://{{TEAM_HOST}}:8080
      service_port: 8080
    runtime_config:
      service_port: 8080
      defense_scope:
        editable_paths:
          - docker/challenge_app.py
        protected_paths:
          - docker/app.py
          - docker/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - PING 必须返回 PONG
`,
		"awd-tcp-length-gate/statement.md":            "TCP checker service.",
		"awd-tcp-length-gate/docker/app.py":           "print('entry')\n",
		"awd-tcp-length-gate/docker/ctf_runtime.py":   "print('runtime')\n",
		"awd-tcp-length-gate/docker/challenge_app.py": "print('challenge')\n",
		"awd-tcp-length-gate/docker/check/check.py":   "print('check')\n",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func readInt64FromAnyForAWDImportTest(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func buildAWDScriptCheckerImportArchive(t *testing.T) []byte {
	t.Helper()
	return buildAWDScriptCheckerImportArchiveWithSlug(t, "script-checker", "print('{\"status\":\"ok\"}')\n")
}

func buildAWDScriptCheckerImportArchiveWithSlug(t *testing.T, slug string, checkerContent string) []byte {
	t.Helper()
	files := map[string]string{
		slug + "/challenge.yml": `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: ` + slug + `
  title: Script Checker AWD
  category: web
  difficulty: hard

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: registry.example.edu/ctf/` + slug + `:v1

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    checker:
      type: script_checker
      config:
        runtime: python3
        entry: docker/check/check.py
        timeout_sec: 10
        args:
          - "{{TARGET_URL}}"
        output: json
    flag_policy:
      mode: dynamic_team
    defense_entry:
      mode: http
    access_config:
      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
    runtime_config:
      service_port: 8080
      defense_scope:
        editable_paths:
          - docker/challenge_app.py
        protected_paths:
          - docker/app.py
          - docker/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`,
		slug + "/statement.md":            "Script checker service.",
		slug + "/docker/check/check.py":   checkerContent,
		slug + "/docker/app.py":           "print('entry')\n",
		slug + "/docker/ctf_runtime.py":   "print('runtime')\n",
		slug + "/docker/challenge_app.py": "print('challenge')\n",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}

func buildAWDMultiFileScriptCheckerImportArchive(t *testing.T) []byte {
	t.Helper()

	files := map[string]string{
		"script-checker-files/challenge.yml": `api_version: v1
kind: challenge

meta:
  mode: awd
  slug: script-checker-files
  title: Script Checker Files AWD
  category: web
  difficulty: hard

content:
  statement: statement.md

flag:
  type: dynamic
  prefix: awd

runtime:
  type: container
  image:
    ref: registry.example.edu/ctf/script-checker-files:v1

extensions:
  awd:
    service_type: web_http
    deployment_mode: single_container
    checker:
      type: script_checker
      config:
        runtime: python3
        entry: docker/check/check.py
        files:
          - docker/check/check.py
          - docker/check/protocol.py
        timeout_sec: 10
        args:
          - "{{TARGET_URL}}"
        output: json
    flag_policy:
      mode: dynamic_team
    defense_entry:
      mode: http
    access_config:
      public_base_url: http://{{TEAM_HOST}}:8080
      service_port: 8080
    runtime_config:
      service_port: 8080
      defense_scope:
        editable_paths:
          - docker/challenge_app.py
        protected_paths:
          - docker/app.py
          - docker/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`,
		"script-checker-files/statement.md":                "Script checker service.",
		"script-checker-files/docker/check/check.py":       "import protocol\nprint(protocol.STATUS)\n",
		"script-checker-files/docker/check/protocol.py":    "STATUS = '{\"status\":\"ok\"}'\n",
		"script-checker-files/docker/check/unused_file.py": "SHOULD_NOT_IMPORT = True\n",
		"script-checker-files/docker/app.py":               "print('entry')\n",
		"script-checker-files/docker/ctf_runtime.py":       "print('runtime')\n",
		"script-checker-files/docker/challenge_app.py":     "print('challenge')\n",
	}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := io.WriteString(fileWriter, content); err != nil {
			t.Fatalf("WriteString(%s) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}
