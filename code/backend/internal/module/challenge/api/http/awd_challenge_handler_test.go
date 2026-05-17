package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

func TestAWDChallengeHandlerListChallenges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAWDChallengeHandler(
		stubAWDChallengeCommandService{},
		stubAWDChallengeQueryService{
			listWithContextFunc: func(ctx context.Context, req challengeqry.ListAWDChallengesInput) (*challengecontracts.AWDChallengePageResp, error) {
				return &challengecontracts.AWDChallengePageResp{
					Items: []*challengecontracts.AWDChallengeResp{
						{ID: 1, Name: "Bank Portal AWD", Slug: "bank-portal-awd"},
					},
					Total: 1,
				}, nil
			},
		},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-challenges", "")
	handler.ListChallenges(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func newJSONTestContext(t *testing.T, method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	return ctx, recorder
}

type stubAWDChallengeCommandService struct {
	listImportsFunc func(ctx context.Context, actorUserID int64) ([]challengecontracts.AWDChallengeImportPreviewResp, error)
	getImportFunc   func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeImportPreviewResp, error)
}

func (stubAWDChallengeCommandService) CreateChallenge(ctx context.Context, actorUserID int64, req challengecmd.CreateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error) {
	return nil, nil
}

func (stubAWDChallengeCommandService) UpdateChallenge(ctx context.Context, id int64, req challengecmd.UpdateAWDChallengeInput) (*challengecontracts.AWDChallengeResp, error) {
	return nil, nil
}

func (stubAWDChallengeCommandService) DeleteChallenge(ctx context.Context, id int64) error {
	return nil
}

func (stubAWDChallengeCommandService) PreviewImport(ctx context.Context, actorUserID int64, fileName string, reader io.Reader) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
	return nil, nil
}

func (s stubAWDChallengeCommandService) ListImports(ctx context.Context, actorUserID int64) ([]challengecontracts.AWDChallengeImportPreviewResp, error) {
	if s.listImportsFunc != nil {
		return s.listImportsFunc(ctx, actorUserID)
	}
	return nil, nil
}

func (s stubAWDChallengeCommandService) GetImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
	if s.getImportFunc != nil {
		return s.getImportFunc(ctx, actorUserID, id)
	}
	return nil, nil
}

func (stubAWDChallengeCommandService) CommitImport(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeResp, error) {
	return nil, nil
}

type stubAWDChallengeQueryService struct {
	listWithContextFunc func(ctx context.Context, req challengeqry.ListAWDChallengesInput) (*challengecontracts.AWDChallengePageResp, error)
}

func (s stubAWDChallengeQueryService) GetChallenge(ctx context.Context, id int64) (*challengecontracts.AWDChallengeResp, error) {
	return nil, nil
}

func (s stubAWDChallengeQueryService) ListChallenges(ctx context.Context, req challengeqry.ListAWDChallengesInput) (*challengecontracts.AWDChallengePageResp, error) {
	if s.listWithContextFunc != nil {
		return s.listWithContextFunc(ctx, req)
	}
	return &challengecontracts.AWDChallengePageResp{}, nil
}

type awdChallengeHandlerContextKey string

func TestAWDChallengeHandlerListImportsPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := awdChallengeHandlerContextKey("list-imports")
	expectedCtxValue := "ctx-list-imports"
	called := false
	handler := NewAWDChallengeHandler(
		stubAWDChallengeCommandService{
			listImportsFunc: func(ctx context.Context, actorUserID int64) ([]challengecontracts.AWDChallengeImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected list-imports ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 2001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				return []challengecontracts.AWDChallengeImportPreviewResp{{ID: "preview-1", Slug: "awd-bank-portal-01"}}, nil
			},
		},
		stubAWDChallengeQueryService{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-challenge-imports", "")
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 2001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.ListImports(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected list-imports command to be called")
	}
}

func TestAWDChallengeHandlerGetImportPropagatesRequestContextToCommandService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctxKey := awdChallengeHandlerContextKey("get-import")
	expectedCtxValue := "ctx-get-import"
	called := false
	handler := NewAWDChallengeHandler(
		stubAWDChallengeCommandService{
			getImportFunc: func(ctx context.Context, actorUserID int64, id string) (*challengecontracts.AWDChallengeImportPreviewResp, error) {
				called = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected get-import ctx value %v, got %v", expectedCtxValue, got)
				}
				if actorUserID != 2001 {
					t.Fatalf("unexpected actor user id: %d", actorUserID)
				}
				if id != "preview-1" {
					t.Fatalf("unexpected import id: %s", id)
				}
				return &challengecontracts.AWDChallengeImportPreviewResp{ID: id, Slug: "awd-bank-portal-01"}, nil
			},
		},
		stubAWDChallengeQueryService{},
	)

	ctx, recorder := newJSONTestContext(t, http.MethodGet, "/admin/awd-challenge-imports/preview-1", "")
	ctx.Params = gin.Params{{Key: "id", Value: "preview-1"}}
	authctx.SetCurrentUser(ctx, authctx.CurrentUser{UserID: 2001})
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxKey, expectedCtxValue))

	handler.GetImport(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected get-import command to be called")
	}
}
