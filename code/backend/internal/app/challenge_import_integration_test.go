package app

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/internal/module/challenge/testsupport"
)

type challengeImportQueryStub struct{}

func (challengeImportQueryStub) GetChallenge(ctx context.Context, id int64) (*dto.ChallengeResp, error) {
	panic("unexpected call")
}

func (challengeImportQueryStub) ListChallenges(ctx context.Context, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeResp], error) {
	panic("unexpected call")
}

func (challengeImportQueryStub) ListPublishedChallenges(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult[*dto.ChallengeListItem], error) {
	panic("unexpected call")
}

func (challengeImportQueryStub) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
	panic("unexpected call")
}

type envelope[T any] struct {
	Code int `json:"code"`
	Data T   `json:"data"`
}

type appChallengeImportDockerBuilder struct{}

func (appChallengeImportDockerBuilder) Build(ctx context.Context, contextPath, dockerfilePath, localRef string) error {
	return nil
}

func (appChallengeImportDockerBuilder) Tag(ctx context.Context, sourceRef, targetRef string) error {
	return nil
}

func (appChallengeImportDockerBuilder) Push(ctx context.Context, targetRef string) error {
	return nil
}

func (appChallengeImportDockerBuilder) Pull(ctx context.Context, targetRef string) error {
	return nil
}

func (appChallengeImportDockerBuilder) Inspect(ctx context.Context, targetRef string) (challengeports.ImageInspectResult, error) {
	return challengeports.ImageInspectResult{Size: 1024}, nil
}

type appChallengeImportRegistryVerifier struct{}

func (appChallengeImportRegistryVerifier) CheckManifest(ctx context.Context, imageRef string) (string, error) {
	return "sha256:app-import", nil
}

func newChallengeImportServiceForAppTest(db *gorm.DB) *challengecmd.ChallengeService {
	repo := challengeinfra.NewRepository(db)
	imageRepo := challengeinfra.NewImageRepository(db)
	imageBuildService := challengecmd.NewImageBuildService(
		imageRepo,
		challengecmd.ImageBuildConfig{Registry: "127.0.0.1:5000"},
		challengecmd.WithImageBuildDockerBuilder(appChallengeImportDockerBuilder{}),
		challengecmd.WithImageBuildRegistryVerifier(appChallengeImportRegistryVerifier{}),
	)
	service := challengecmd.NewChallengeService(db, repo, imageRepo, nil, nil, nil, challengecmd.SelfCheckConfig{}, zap.NewNop())
	service.SetImageBuildService(imageBuildService)
	return service
}

func TestChallengeImportPreviewAndCommitFlow(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	body, contentType := buildChallengeImportMultipart(t)
	previewRequest := httptest.NewRequest(nethttp.MethodPost, "/imports", body)
	previewRequest.Header.Set("Content-Type", contentType)
	previewRequest.Header.Set("X-Test-User-ID", "1001")
	previewRecorder := httptest.NewRecorder()
	router.ServeHTTP(previewRecorder, previewRequest)

	if previewRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("preview status = %d, body = %s", previewRecorder.Code, previewRecorder.Body.String())
	}

	var previewEnvelope envelope[dto.ChallengeImportPreviewResp]
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &previewEnvelope); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if previewEnvelope.Data.ID == "" {
		t.Fatal("expected preview id")
	}
	if previewEnvelope.Data.Title != "SQL Injection 101" {
		t.Fatalf("unexpected preview title: %s", previewEnvelope.Data.Title)
	}

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest(nethttp.MethodGet, "/imports/"+previewEnvelope.Data.ID, nil)
	getRequest.Header.Set("X-Test-User-ID", "1001")
	router.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != nethttp.StatusOK {
		t.Fatalf("get preview status = %d, body = %s", getRecorder.Code, getRecorder.Body.String())
	}

	commitRecorder := httptest.NewRecorder()
	commitRequest := httptest.NewRequest(nethttp.MethodPost, "/imports/"+previewEnvelope.Data.ID+"/commit", nil)
	commitRequest.Header.Set("X-Test-User-ID", "1001")
	router.ServeHTTP(commitRecorder, commitRequest)
	if commitRecorder.Code != nethttp.StatusOK {
		t.Fatalf("commit status = %d, body = %s", commitRecorder.Code, commitRecorder.Body.String())
	}

	var commitEnvelope envelope[dto.ChallengeImportCommitResp]
	if err := json.Unmarshal(commitRecorder.Body.Bytes(), &commitEnvelope); err != nil {
		t.Fatalf("decode commit response: %v", err)
	}
	if commitEnvelope.Data.Challenge == nil {
		t.Fatal("expected imported challenge response")
	}
	if commitEnvelope.Data.Challenge.Title != "SQL Injection 101" {
		t.Fatalf("unexpected imported challenge title: %s", commitEnvelope.Data.Challenge.Title)
	}
	if !strings.HasPrefix(commitEnvelope.Data.Challenge.AttachmentURL, "/api/v1/challenges/attachments/imports/") {
		t.Fatalf("unexpected attachment url: %s", commitEnvelope.Data.Challenge.AttachmentURL)
	}
	if !strings.HasSuffix(commitEnvelope.Data.Challenge.AttachmentURL, ".zip") {
		t.Fatalf("expected bundled attachment zip, got %s", commitEnvelope.Data.Challenge.AttachmentURL)
	}
}

func TestChallengeImportCommitUpsertsByPackageSlug(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	legacyChallenge := model.Challenge{
		Title:       "SQL Injection 101",
		Description: "legacy",
		Category:    "web",
		Difficulty:  "easy",
		Points:      50,
		Status:      model.ChallengeStatusDraft,
	}
	if err := db.Create(&legacyChallenge).Error; err != nil {
		t.Fatalf("seed legacy challenge: %v", err)
	}

	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	firstCommit := previewAndCommitChallengeImport(
		t,
		router,
		buildChallengeImportArchiveForSlug(t, "web-sqli-101", "SQL Injection 101", 100),
	)
	if firstCommit.Challenge == nil {
		t.Fatal("expected first imported challenge response")
	}
	if firstCommit.Challenge.ID != legacyChallenge.ID {
		t.Fatalf("expected legacy challenge to be reused, got id=%d", firstCommit.Challenge.ID)
	}

	var packageSlug string
	if err := db.Raw("SELECT package_slug FROM challenges WHERE id = ?", legacyChallenge.ID).Scan(&packageSlug).Error; err != nil {
		t.Fatalf("query package_slug after first import: %v", err)
	}
	if packageSlug != "web-sqli-101" {
		t.Fatalf("expected package_slug to be persisted, got %q", packageSlug)
	}

	secondCommit := previewAndCommitChallengeImport(
		t,
		router,
		buildChallengeImportArchiveForSlug(t, "web-sqli-101", "SQL Injection 102", 200),
	)
	if secondCommit.Challenge == nil {
		t.Fatal("expected second imported challenge response")
	}
	if secondCommit.Challenge.ID != legacyChallenge.ID {
		t.Fatalf("expected slug upsert to update same challenge, got id=%d", secondCommit.Challenge.ID)
	}
	if secondCommit.Challenge.Title != "SQL Injection 102" {
		t.Fatalf("expected updated title, got %q", secondCommit.Challenge.Title)
	}
	if secondCommit.Challenge.Points != 200 {
		t.Fatalf("expected updated points, got %d", secondCommit.Challenge.Points)
	}

	var count int64
	if err := db.Model(&model.Challenge{}).Count(&count).Error; err != nil {
		t.Fatalf("count challenges: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 challenge after slug upsert, got %d", count)
	}
}

func TestChallengeImportGetRejectsDifferentAdmin(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	body, contentType := buildChallengeImportMultipart(t)
	previewRequest := httptest.NewRequest(nethttp.MethodPost, "/imports", body)
	previewRequest.Header.Set("Content-Type", contentType)
	previewRequest.Header.Set("X-Test-User-ID", "1001")
	previewRecorder := httptest.NewRecorder()
	router.ServeHTTP(previewRecorder, previewRequest)
	if previewRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("preview status = %d, body = %s", previewRecorder.Code, previewRecorder.Body.String())
	}

	var previewEnvelope envelope[dto.ChallengeImportPreviewResp]
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &previewEnvelope); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest(nethttp.MethodGet, "/imports/"+previewEnvelope.Data.ID, nil)
	getRequest.Header.Set("X-Test-User-ID", "2002")
	router.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != nethttp.StatusForbidden {
		t.Fatalf("expected forbidden get preview, status = %d, body = %s", getRecorder.Code, getRecorder.Body.String())
	}
}

func TestChallengeImportCommitSupportsRegexFlag(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	commit := previewAndCommitChallengeImport(
		t,
		router,
		buildChallengeImportArchiveWithFlagConfig(t, "regex", `^flag\{import-[0-9]{2}\}$`, "flag"),
	)
	if commit.Challenge == nil {
		t.Fatal("expected imported challenge response")
	}

	var stored model.Challenge
	if err := db.First(&stored, commit.Challenge.ID).Error; err != nil {
		t.Fatalf("load imported challenge: %v", err)
	}
	if stored.FlagType != model.FlagTypeRegex || stored.FlagRegex != `^flag\{import-[0-9]{2}\}$` {
		t.Fatalf("expected regex flag persisted, got %+v", stored)
	}
}

func TestChallengeImportCommitSupportsManualReviewFlag(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	commit := previewAndCommitChallengeImport(
		t,
		router,
		buildChallengeImportArchiveWithFlagConfig(t, "manual_review", "", ""),
	)
	if commit.Challenge == nil {
		t.Fatal("expected imported challenge response")
	}

	var stored model.Challenge
	if err := db.First(&stored, commit.Challenge.ID).Error; err != nil {
		t.Fatalf("load imported challenge: %v", err)
	}
	if stored.FlagType != model.FlagTypeManualReview {
		t.Fatalf("expected manual review flag persisted, got %+v", stored)
	}
}

func TestChallengeImportPreviewRejectsArchiveWithTooManyFiles(t *testing.T) {
	t.Setenv("CHALLENGE_IMPORT_PREVIEW_DIR", t.TempDir())
	t.Setenv("CHALLENGE_ATTACHMENT_STORAGE_DIR", t.TempDir())

	db := testsupport.SetupTestDB(t)
	service := newChallengeImportServiceForAppTest(db)
	router := buildChallengeImportRouter(service)

	body, contentType := buildChallengeImportMultipartFromArchive(t, buildChallengeImportArchiveWithTooManyFiles(t))
	request := httptest.NewRequest(nethttp.MethodPost, "/imports", body)
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("X-Test-User-ID", "1001")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected bad request for oversized archive, status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func buildChallengeImportRouter(service *challengecmd.ChallengeService) *gin.Engine {
	handler := challengehttp.NewHandler(service, challengeImportQueryStub{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		userID := int64(1001)
		if raw := strings.TrimSpace(c.GetHeader("X-Test-User-ID")); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err == nil {
				userID = parsed
			}
		}
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: userID, Role: "admin"})
		c.Next()
	})
	router.POST("/imports", handler.PreviewChallengeImport)
	router.GET("/imports/:id", handler.GetChallengeImport)
	router.POST("/imports/:id/commit", handler.CommitChallengeImport)
	return router
}

func buildChallengeImportMultipart(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	return buildChallengeImportMultipartFromArchive(t, buildChallengeImportArchive(t))
}

func buildChallengeImportMultipartFromArchive(t *testing.T, archiveBytes []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "web-sqli-101.zip")
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

func buildChallengeImportArchive(t *testing.T) []byte {
	t.Helper()

	return buildChallengeImportArchiveForSlug(t, "web-sqli-101", "SQL Injection 101", 100)
}

func buildChallengeImportArchiveWithFlagConfig(t *testing.T, flagType, flagValue, flagPrefix string) []byte {
	t.Helper()

	return buildChallengeImportArchiveForSlugAndFlag(t, "web-sqli-101", "SQL Injection 101", 100, flagType, flagValue, flagPrefix)
}

func buildChallengeImportArchiveForSlug(t *testing.T, slug, title string, points int) []byte {
	t.Helper()

	return buildChallengeImportArchiveForSlugAndFlag(t, slug, title, points, "static", "flag{sqli_101}", "flag")
}

func buildChallengeImportArchiveForSlugAndFlag(t *testing.T, slug, title string, points int, flagType, flagValue, flagPrefix string) []byte {
	t.Helper()

	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)
	flagBlock := "flag:\n  type: " + flagType + "\n"
	if flagValue != "" {
		flagBlock += "  value: " + flagValue + "\n"
	}
	if flagPrefix != "" {
		flagBlock += "  prefix: " + flagPrefix + "\n"
	}
	files := map[string]string{
		slug + `/challenge.yml`: `
api_version: v1
kind: challenge
meta:
  slug: ` + slug + `
  title: ` + title + `
  category: web
  difficulty: easy
  points: ` + fmt.Sprintf("%d", points) + `
content:
  statement: statement.md
  attachments:
    - path: attachments/web-sqli-101.zip
      name: web-sqli-101.zip
    - path: attachments/readme.txt
      name: readme.txt
` + flagBlock + `
hints:
  - level: 1
    title: Hint 1
    content: 从登录参数开始看
runtime:
  type: container
  image:
    ref: ctf/web-sqli-101:latest
extensions:
  topology:
    source: docker/topology.yml
    enabled: false
`,
		slug + `/statement.md`:                 "# SQLi 101\n\nFind the bypass.",
		slug + `/attachments/web-sqli-101.zip`: "archive-bytes",
		slug + `/attachments/readme.txt`:       "remember to inspect login params",
	}

	for name, content := range files {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close zip archive: %v", err)
	}
	return buffer.Bytes()
}

func previewAndCommitChallengeImport(
	t *testing.T,
	router *gin.Engine,
	archiveBytes []byte,
) dto.ChallengeImportCommitResp {
	t.Helper()

	body, contentType := buildChallengeImportMultipartFromArchive(t, archiveBytes)
	previewRequest := httptest.NewRequest(nethttp.MethodPost, "/imports", body)
	previewRequest.Header.Set("Content-Type", contentType)
	previewRequest.Header.Set("X-Test-User-ID", "1001")
	previewRecorder := httptest.NewRecorder()
	router.ServeHTTP(previewRecorder, previewRequest)
	if previewRecorder.Code != nethttp.StatusCreated {
		t.Fatalf("preview status = %d, body = %s", previewRecorder.Code, previewRecorder.Body.String())
	}

	var previewEnvelope envelope[dto.ChallengeImportPreviewResp]
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &previewEnvelope); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}

	commitRecorder := httptest.NewRecorder()
	commitRequest := httptest.NewRequest(nethttp.MethodPost, "/imports/"+previewEnvelope.Data.ID+"/commit", nil)
	commitRequest.Header.Set("X-Test-User-ID", "1001")
	router.ServeHTTP(commitRecorder, commitRequest)
	if commitRecorder.Code != nethttp.StatusOK {
		t.Fatalf("commit status = %d, body = %s", commitRecorder.Code, commitRecorder.Body.String())
	}

	var commitEnvelope envelope[dto.ChallengeImportCommitResp]
	if err := json.Unmarshal(commitRecorder.Body.Bytes(), &commitEnvelope); err != nil {
		t.Fatalf("decode commit response: %v", err)
	}
	return commitEnvelope.Data
}

func buildChallengeImportArchiveWithTooManyFiles(t *testing.T) []byte {
	t.Helper()

	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)
	for index := 0; index < 129; index++ {
		name := fmt.Sprintf("oversized/file-%03d.txt", index)
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := entry.Write([]byte("x")); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close zip archive: %v", err)
	}
	return buffer.Bytes()
}
