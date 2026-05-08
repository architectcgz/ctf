package runtime

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	"ctf-platform/internal/module/challenge/testsupport"
)

type importEnvelope[T any] struct {
	Code int `json:"code"`
	Data T   `json:"data"`
}

func TestBuildWiresChallengeImportImageBuildService(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())
	t.Setenv("CHALLENGE_IMAGE_BUILD_SOURCE_DIR", t.TempDir())

	deps, db := newChallengeRuntimeImportDeps(t)
	module, err := Build(deps)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	closeChallengeRuntimeModule(t, module)

	router := newChallengeImportTestRouter(module)
	body, contentType := buildMultipartArchive(t, "web-platform-build.zip", buildChallengePlatformBuildArchive(t))

	previewRecorder := httptest.NewRecorder()
	previewRequest := httptest.NewRequest(nethttp.MethodPost, "/imports", body)
	previewRequest.Header.Set("Content-Type", contentType)
	router.ServeHTTP(previewRecorder, previewRequest)
	if previewRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("preview status = %d, body = %s", previewRecorder.Code, previewRecorder.Body.String())
	}

	var preview importEnvelope[dto.ChallengeImportPreviewResp]
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if preview.Data.ImageDelivery.SourceType != model.ImageSourceTypePlatformBuild {
		t.Fatalf("preview source_type = %q, want %q", preview.Data.ImageDelivery.SourceType, model.ImageSourceTypePlatformBuild)
	}
	if preview.Data.ImageDelivery.TargetImageRef != "127.0.0.1:5000/jeopardy/web-platform-build:v1" {
		t.Fatalf("preview target_image_ref = %q", preview.Data.ImageDelivery.TargetImageRef)
	}
	if preview.Data.ImageDelivery.BuildStatus != model.ImageStatusPending {
		t.Fatalf("preview build_status = %q, want pending", preview.Data.ImageDelivery.BuildStatus)
	}

	commitRecorder := httptest.NewRecorder()
	commitRequest := httptest.NewRequest(nethttp.MethodPost, "/imports/"+preview.Data.ID+"/commit", nil)
	router.ServeHTTP(commitRecorder, commitRequest)
	if commitRecorder.Code != nethttp.StatusOK {
		t.Fatalf("commit status = %d, body = %s", commitRecorder.Code, commitRecorder.Body.String())
	}

	var commit importEnvelope[dto.ChallengeImportCommitResp]
	if err := json.Unmarshal(commitRecorder.Body.Bytes(), &commit); err != nil {
		t.Fatalf("decode commit response: %v", err)
	}
	if commit.Data.Challenge == nil || commit.Data.Challenge.ImageID <= 0 {
		t.Fatalf("unexpected commit response: %+v", commit.Data)
	}

	imageRepo := challengeinfra.NewImageRepository(db)
	image, err := imageRepo.FindByID(context.Background(), commit.Data.Challenge.ImageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "127.0.0.1:5000/jeopardy/web-platform-build" || image.Tag != "v1" {
		t.Fatalf("unexpected stored image: %+v", image)
	}
	if image.BuildJobID == nil || image.Status != model.ImageStatusPending {
		t.Fatalf("expected pending platform build image, got %+v", image)
	}
}

func TestBuildWiresAWDImportImageBuildService(t *testing.T) {
	t.Setenv("AWD_CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_IMAGE_BUILD_SOURCE_DIR", t.TempDir())

	deps, db := newChallengeRuntimeImportDeps(t)
	module, err := Build(deps)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	closeChallengeRuntimeModule(t, module)

	router := newAWDImportTestRouter(module)
	body, contentType := buildMultipartArchive(t, "awd-platform-build.zip", buildAWDPlatformBuildArchive(t))

	previewRecorder := httptest.NewRecorder()
	previewRequest := httptest.NewRequest(nethttp.MethodPost, "/awd-imports", body)
	previewRequest.Header.Set("Content-Type", contentType)
	router.ServeHTTP(previewRecorder, previewRequest)
	if previewRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("preview status = %d, body = %s", previewRecorder.Code, previewRecorder.Body.String())
	}

	var preview importEnvelope[dto.AWDChallengeImportPreviewResp]
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if preview.Data.ImageDelivery.SourceType != model.ImageSourceTypePlatformBuild {
		t.Fatalf("preview source_type = %q, want %q", preview.Data.ImageDelivery.SourceType, model.ImageSourceTypePlatformBuild)
	}
	if preview.Data.ImageDelivery.TargetImageRef != "127.0.0.1:5000/awd/awd-platform-build:c1" {
		t.Fatalf("preview target_image_ref = %q", preview.Data.ImageDelivery.TargetImageRef)
	}
	if preview.Data.ImageDelivery.BuildStatus != model.ImageStatusPending {
		t.Fatalf("preview build_status = %q, want pending", preview.Data.ImageDelivery.BuildStatus)
	}

	commitRecorder := httptest.NewRecorder()
	commitRequest := httptest.NewRequest(nethttp.MethodPost, "/awd-imports/"+preview.Data.ID+"/commit", nil)
	router.ServeHTTP(commitRecorder, commitRequest)
	if commitRecorder.Code != nethttp.StatusOK {
		t.Fatalf("commit status = %d, body = %s", commitRecorder.Code, commitRecorder.Body.String())
	}

	var commit importEnvelope[dto.AWDChallengeImportCommitResp]
	if err := json.Unmarshal(commitRecorder.Body.Bytes(), &commit); err != nil {
		t.Fatalf("decode commit response: %v", err)
	}
	if commit.Data.Challenge == nil {
		t.Fatalf("unexpected commit response: %+v", commit.Data)
	}
	imageID := readInt64FromAny(commit.Data.Challenge.RuntimeConfig["image_id"])
	if imageID <= 0 {
		t.Fatalf("expected runtime_config.image_id, got %+v", commit.Data.Challenge.RuntimeConfig)
	}

	imageRepo := challengeinfra.NewImageRepository(db)
	image, err := imageRepo.FindByID(context.Background(), imageID)
	if err != nil {
		t.Fatalf("FindByID(image) error = %v", err)
	}
	if image.Name != "127.0.0.1:5000/awd/awd-platform-build" || image.Tag != "c1" {
		t.Fatalf("unexpected stored image: %+v", image)
	}
	if image.BuildJobID == nil || image.Status != model.ImageStatusPending {
		t.Fatalf("expected pending awd platform build image, got %+v", image)
	}
}

func newChallengeRuntimeImportDeps(t *testing.T) (Deps, *gorm.DB) {
	t.Helper()

	db := testsupport.SetupTestDB(t)
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = cache.Close()
	})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	appCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Container: config.ContainerConfig{
			FlagGlobalSecret: "integration-secret-123456789012345",
			Registry: config.ContainerRegistryConfig{
				Enabled:          true,
				BuildEnabled:     false,
				Scheme:           "http",
				Server:           "127.0.0.1:5000",
				Username:         "ctf",
				Password:         "123456",
				BuildTimeout:     10 * time.Minute,
				BuildConcurrency: 1,
			},
		},
		Challenge: config.ChallengeConfig{
			PublishCheck: config.ChallengePublishCheckConfig{
				Enabled: false,
			},
		},
	}

	return Deps{
		AppContext: appCtx,
		Config:     cfg,
		Logger:     zap.NewNop(),
		DB:         db,
		Cache:      cache,
	}, db
}

func closeChallengeRuntimeModule(t *testing.T, module *Module) {
	t.Helper()
	if module == nil || module.BackgroundTasks == nil {
		return
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := module.BackgroundTasks.Close(ctx); err != nil {
			t.Fatalf("close background tasks: %v", err)
		}
	})
}

func newChallengeImportTestRouter(module *Module) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(importTestAuthMiddleware())
	router.POST("/imports", module.Handler.PreviewChallengeImport)
	router.POST("/imports/:id/commit", module.Handler.CommitChallengeImport)
	return router
}

func newAWDImportTestRouter(module *Module) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(importTestAuthMiddleware())
	router.POST("/awd-imports", module.AWDChallengeHandler.PreviewImport)
	router.POST("/awd-imports/:id/commit", module.AWDChallengeHandler.CommitImport)
	return router
}

func importTestAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := int64(1001)
		if raw := strings.TrimSpace(c.GetHeader("X-Test-User-ID")); raw != "" {
			if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
				userID = parsed
			}
		}
		authctx.SetCurrentUser(c, authctx.CurrentUser{
			UserID:   userID,
			Username: "tester",
			Role:     model.RoleAdmin,
		})
		c.Next()
	}
}

func buildMultipartArchive(t *testing.T, fileName string, archiveBytes []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(archiveBytes); err != nil {
		t.Fatalf("write archive body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func buildChallengePlatformBuildArchive(t *testing.T) []byte {
	t.Helper()

	files := map[string]string{
		"web-platform-build/challenge.yml": `api_version: v1
kind: challenge

meta:
  slug: web-platform-build
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
`,
		"web-platform-build/statement.md":      "platform build statement",
		"web-platform-build/docker/Dockerfile": "FROM nginx:1.27-alpine\n",
	}
	return buildImportArchive(t, files)
}

func buildAWDPlatformBuildArchive(t *testing.T) []byte {
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
      defense_workspace:
        entry_mode: ssh
        seed_root: docker/workspace
        workspace_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
          - docker/workspace/data
        writable_roots:
          - docker/workspace/src
          - docker/workspace/templates
          - docker/workspace/static
        readonly_roots:
          - docker/workspace/data
        runtime_mounts:
          - source: docker/workspace/src
            target: /workspace/src
            mode: rw
          - source: docker/workspace/templates
            target: /workspace/templates
            mode: rw
          - source: docker/workspace/static
            target: /workspace/static
            mode: rw
          - source: docker/workspace/data
            target: /workspace/data
            mode: ro
      defense_scope:
        protected_paths:
          - docker/runtime/app.py
          - docker/runtime/ctf_runtime.py
          - docker/check/check.py
          - challenge.yml
        service_contracts:
          - /health 必须返回 200
`,
		"awd-platform-build/statement.md":                          "平台构建 AWD 服务。",
		"awd-platform-build/docker/runtime/Dockerfile":             "FROM python:3.12-alpine\nWORKDIR /app\nCOPY runtime /app/runtime\n",
		"awd-platform-build/docker/runtime/app.py":                 "print('entry')\n",
		"awd-platform-build/docker/runtime/ctf_runtime.py":         "print('runtime')\n",
		"awd-platform-build/docker/check/check.py":                 "print('check')\n",
		"awd-platform-build/docker/workspace/src/app.py":           "print('workspace entry')\n",
		"awd-platform-build/docker/workspace/src/service.py":       "print('service logic')\n",
		"awd-platform-build/docker/workspace/templates/index.html": "<h1>workspace</h1>\n",
		"awd-platform-build/docker/workspace/static/site.css":      "body { color: black; }\n",
		"awd-platform-build/docker/workspace/data/seed.txt":        "seed\n",
	}
	return buildImportArchive(t, files)
}

func buildImportArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

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

func readInt64FromAny(value any) int64 {
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
