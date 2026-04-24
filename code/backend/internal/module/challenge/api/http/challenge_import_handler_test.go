package http

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
)

type challengeImportHandlerCommandStub struct {
	listChallengeImportsWithContextFn func(ctx context.Context, actorUserID int64) ([]dto.ChallengeImportPreviewResp, error)
	getChallengeImportWithContextFn   func(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error)
}

func (s challengeImportHandlerCommandStub) CreateChallenge(actorUserID int64, req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) CreateChallengeWithContext(ctx context.Context, actorUserID int64, req *dto.CreateChallengeReq) (*dto.ChallengeResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) UpdateChallenge(id int64, req *dto.UpdateChallengeReq) error {
	return nil
}

func (s challengeImportHandlerCommandStub) UpdateChallengeWithContext(ctx context.Context, id int64, req *dto.UpdateChallengeReq) error {
	return nil
}

func (s challengeImportHandlerCommandStub) DeleteChallenge(id int64) error {
	return nil
}

func (s challengeImportHandlerCommandStub) DeleteChallengeWithContext(ctx context.Context, id int64) error {
	return nil
}

func (s challengeImportHandlerCommandStub) RequestPublishCheck(actorUserID, id int64) (*dto.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) RequestPublishCheckWithContext(ctx context.Context, actorUserID, id int64) (*dto.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetLatestPublishCheck(id int64) (*dto.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetLatestPublishCheckWithContext(ctx context.Context, id int64) (*dto.ChallengePublishCheckJobResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) SelfCheckChallenge(id int64) (*dto.ChallengeSelfCheckResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) SelfCheckChallengeWithContext(ctx context.Context, id int64) (*dto.ChallengeSelfCheckResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) PreviewChallengeImport(actorUserID int64, fileName string, reader io.Reader) (*dto.ChallengeImportPreviewResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) PreviewChallengeImportWithContext(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*dto.ChallengeImportPreviewResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) ListChallengeImports(actorUserID int64) ([]dto.ChallengeImportPreviewResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) ListChallengeImportsWithContext(ctx context.Context, actorUserID int64) ([]dto.ChallengeImportPreviewResp, error) {
	if s.listChallengeImportsWithContextFn != nil {
		return s.listChallengeImportsWithContextFn(ctx, actorUserID)
	}
	return s.ListChallengeImports(actorUserID)
}

func (s challengeImportHandlerCommandStub) GetChallengeImport(actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) GetChallengeImportWithContext(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error) {
	if s.getChallengeImportWithContextFn != nil {
		return s.getChallengeImportWithContextFn(ctx, actorUserID, id)
	}
	return s.GetChallengeImport(actorUserID, id)
}

func (s challengeImportHandlerCommandStub) CommitChallengeImport(actorUserID int64, id string) (*dto.ChallengeResp, error) {
	return nil, nil
}

func (s challengeImportHandlerCommandStub) CommitChallengeImportWithContext(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeResp, error) {
	return nil, nil
}

type challengeImportHandlerQueryStub struct{}

func (challengeImportHandlerQueryStub) GetChallenge(ctx context.Context, id int64) (*dto.ChallengeResp, error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) ListChallenges(ctx context.Context, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) ListPublishedChallenges(ctx context.Context, userID int64, query *dto.ChallengeQuery) (*dto.PageResult, error) {
	return nil, nil
}

func (challengeImportHandlerQueryStub) GetPublishedChallenge(ctx context.Context, userID, challengeID int64) (*dto.ChallengeDetailResp, error) {
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
			listChallengeImportsWithContextFn: func(ctx context.Context, actorUserID int64) ([]dto.ChallengeImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected list-imports ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 1001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				return []dto.ChallengeImportPreviewResp{{ID: "preview-1", Slug: "web-source-audit-double-wrap-01"}}, nil
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
			getChallengeImportWithContextFn: func(ctx context.Context, actorUserID int64, id string) (*dto.ChallengeImportPreviewResp, error) {
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
				return &dto.ChallengeImportPreviewResp{ID: id, Slug: "web-source-audit-double-wrap-01"}, nil
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
