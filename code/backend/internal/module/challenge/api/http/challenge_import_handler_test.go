package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

type challengeImportHandlerCommandStub struct {
	previewChallengeImportFn func(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error)
	listChallengeImportsFn   func(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error)
	getChallengeImportFn     func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error)
	commitChallengeImportFn  func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error)
}

func (s challengeImportHandlerCommandStub) CreateChallenge(ctx context.Context, actorUserID int64, req challengecmd.CreateChallengeInput) (*challengecontracts.ChallengeResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) UpdateChallenge(ctx context.Context, id int64, req challengecmd.UpdateChallengeInput) error {
	return nil
}

func (s challengeImportHandlerCommandStub) DeleteChallenge(ctx context.Context, id int64) error {
	return nil
}

func (s challengeImportHandlerCommandStub) RequestPublishCheck(ctx context.Context, actorUserID, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetLatestPublishCheck(ctx context.Context, id int64) (*challengecontracts.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) SelfCheckChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeSelfCheckResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) PreviewChallengeImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error) {
	if s.previewChallengeImportFn != nil {
		return s.previewChallengeImportFn(ctx, actorUserID, fileName, reader)
	}
	return nil, nil
}

func (s challengeImportHandlerCommandStub) ListChallengeImports(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error) {
	if s.listChallengeImportsFn != nil {
		return s.listChallengeImportsFn(ctx, actorUserID)
	}
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error) {
	if s.getChallengeImportFn != nil {
		return s.getChallengeImportFn(ctx, actorUserID, id)
	}
	return nil, nil
}

func (s challengeImportHandlerCommandStub) CommitChallengeImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error) {
	if s.commitChallengeImportFn != nil {
		return s.commitChallengeImportFn(ctx, actorUserID, id)
	}
	return nil, nil
}

func (s challengeImportHandlerCommandStub) ExportChallengePackage(ctx context.Context, actorUserID int64, challengeID int64) (*challengecontracts.ChallengePackageExportResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetChallengePackageExport(ctx context.Context, challengeID int64, revisionID *int64) (*challengecontracts.ChallengePackageExportResp, error) {
	return nil, nil
}

type challengeImportHandlerQueryStub struct{}

func (challengeImportHandlerQueryStub) GetChallenge(ctx context.Context, id int64) (*challengecontracts.ChallengeResp, error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) ListChallenges(ctx context.Context, query *challengecontracts.ChallengeQuery) (*challengecontracts.PageResult[*challengecontracts.ChallengeResp], error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) ListPublishedChallenges(ctx context.Context, userID int64, query *challengecontracts.ChallengeQuery) (*challengecontracts.PageResult[*challengecontracts.ChallengeListItem], error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*challengecontracts.ChallengeDetailResp, error) {
	return nil, nil
}

type challengeImportHandlerContextKey string

func TestHandlerPreviewChallengeImportUsesPackageDeliveryFacade(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	handler := NewHandler(
		challengeImportHandlerCommandStub{
			previewChallengeImportFn: func(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.ChallengeImportPreviewResp, error) {
				called = true
				if actorUserID != 1001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				if fileName != "challenge.zip" {
					t.Fatalf("unexpected file name: %s", fileName)
				}
				content, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("read upload: %v", err)
				}
				if string(content) != "zip" {
					t.Fatalf("unexpected upload content: %q", string(content))
				}
				return &challengecontracts.ChallengeImportPreviewResp{
					ID:        "preview-1",
					FileName:  fileName,
					Slug:      "web-source-audit",
					Title:     "Web Source Audit",
					CreatedAt: time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC),
				}, nil
			},
		},
		challengeImportHandlerQueryStub{},
	)

	ctx, recorder := newMultipartFileTestContext(t, http.MethodPost, "/admin/challenge-imports", "file", "challenge.zip", "zip")
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 1001})

	handler.PreviewChallengeImport(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected package delivery facade to call preview delegate")
	}
}

func TestHandlerCommitChallengeImportUsesPackageDeliveryFacade(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	handler := NewHandler(
		challengeImportHandlerCommandStub{
			commitChallengeImportFn: func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeResp, error) {
				called = true
				if actorUserID != 1001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				if id != "preview-1" {
					t.Fatalf("unexpected import id: %s", id)
				}
				return &challengecontracts.ChallengeResp{ID: 42, Title: "Imported"}, nil
			},
		},
		challengeImportHandlerQueryStub{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodPost, "/admin/challenge-imports/preview-1/commit", "")
	ctx.Params = gin.Params{{Key: "id", Value: "preview-1"}}
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 1001})

	handler.CommitChallengeImport(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected package delivery facade to call commit delegate")
	}
}

func TestHandlerListChallengeImportsPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := challengeImportHandlerContextKey("list-imports")
	expectedCtxValue := "ctx-list-imports"
	called := false
	handler := NewHandler(
		challengeImportHandlerCommandStub{
			listChallengeImportsFn: func(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected list-imports ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 1001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				return []challengecontracts.ChallengeImportPreviewResp{{ID: "preview-1", Slug: "web-source-audit-double-wrap-01"}}, nil
			},
		},
		challengeImportHandlerQueryStub{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/challenge-imports", "")
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 1001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.ListChallengeImports(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected list challenge imports command to be called")
	}
}

func TestHandlerGetChallengeImportPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := challengeImportHandlerContextKey("get-import")
	expectedCtxValue := "ctx-get-import"
	called := false
	handler := NewHandler(
		challengeImportHandlerCommandStub{
			getChallengeImportFn: func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected get-import ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 1001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				if id != "preview-1" {
					t.Fatalf("unexpected import id: %s", id)
				}
				return &challengecontracts.ChallengeImportPreviewResp{ID: id, Slug: "web-source-audit-double-wrap-01"}, nil
			},
		},
		challengeImportHandlerQueryStub{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/challenge-imports/preview-1", "")
	ctx.Params = gin.Params{{Key: "id", Value: "preview-1"}}
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 1001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.GetChallengeImport(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected get challenge import command to be called")
	}
}

func newMultipartFileTestContext(t *testing.T, method, target, fieldName, fileName, content string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	return ctx, recorder
}
