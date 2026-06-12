package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestHandlerDownloadAttachmentUsesInjectedStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	handler := newChallengeImportHandlerForTest(challengeImportHandlerCommandStub{
		openAttachmentFn: func(ctx context.Context, relativePath string) (*challengeports.ChallengeAttachmentDownload, error) {
			called = true
			if relativePath != "imports/demo/readme.txt" {
				t.Fatalf("unexpected relative path: %s", relativePath)
			}
			return &challengeports.ChallengeAttachmentDownload{
				FileName: "readme.txt",
				Reader:   io.NopCloser(stringsReader("downloaded attachment")),
				Size:     int64(len("downloaded attachment")),
			}, nil
		},
	})

	router := gin.New()
	router.GET("/attachments/*path", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001})
		handler.DownloadAttachment(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/attachments/imports/demo/readme.txt", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !called {
		t.Fatal("expected attachment store to be called")
	}
	if resp.Body.String() != "downloaded attachment" {
		t.Fatalf("body = %q", resp.Body.String())
	}
}

func TestHandlerDownloadAttachmentReturnsNotFoundFromStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newChallengeImportHandlerForTest(challengeImportHandlerCommandStub{
		openAttachmentFn: func(ctx context.Context, relativePath string) (*challengeports.ChallengeAttachmentDownload, error) {
			return nil, apperror.ErrNotFound
		},
	})

	router := gin.New()
	router.GET("/attachments/*path", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001})
		handler.DownloadAttachment(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/attachments/imports/demo/missing.txt", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestHandlerDownloadAttachmentRejectsTraversalBeforeStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	handler := newChallengeImportHandlerForTest(challengeImportHandlerCommandStub{
		openAttachmentFn: func(ctx context.Context, relativePath string) (*challengeports.ChallengeAttachmentDownload, error) {
			called = true
			return nil, nil
		},
	})

	router := gin.New()
	router.GET("/attachments/*path", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001})
		handler.DownloadAttachment(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/attachments/imports/../secret.txt", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
	if called {
		t.Fatal("expected traversal request to be rejected before store call")
	}
}

type staticAttachmentReader struct {
	remaining []byte
}

func stringsReader(value string) *staticAttachmentReader {
	return &staticAttachmentReader{remaining: []byte(value)}
}

func (r *staticAttachmentReader) Read(p []byte) (int, error) {
	if len(r.remaining) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.remaining)
	r.remaining = r.remaining[n:]
	return n, nil
}
