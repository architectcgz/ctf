package http

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

type challengeImportHandlerCommandStub struct {
	listChallengeImportsFn func(ctx context.Context, actorUserID int64) ([]challengecontracts.ChallengeImportPreviewResp, error)
	getChallengeImportFn   func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.ChallengeImportPreviewResp, error)
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
