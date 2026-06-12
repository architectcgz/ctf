package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	assessmentcommands "ctf-platform/internal/module/assessment/application/commands"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"ctf-platform/internal/apperror"
)

type stubReportService struct {
	getDownloadFn func(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error)
}

func (s stubReportService) CreatePersonalReport(context.Context, int64, assessmentcommands.CreatePersonalReportInput) (*assessmentcommands.ReportExportData, error) {
	return nil, nil
}

func (s stubReportService) CreateClassReport(context.Context, int64, assessmentcommands.CreateClassReportInput) (*assessmentcommands.ReportExportData, error) {
	return nil, nil
}

func (s stubReportService) CreateContestExport(context.Context, int64, int64, assessmentcommands.CreateContestExportInput) (*assessmentcommands.ReportExportData, error) {
	return nil, nil
}

func (s stubReportService) CreateStudentReviewArchive(context.Context, int64, int64, assessmentcommands.CreateStudentReviewArchiveInput) (*assessmentcommands.ReportExportData, error) {
	return nil, nil
}

func (s stubReportService) GetStudentReviewArchive(context.Context, int64, int64) (*assessmentcommands.ReviewArchiveData, error) {
	return nil, nil
}

func (s stubReportService) GetDownload(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error) {
	if s.getDownloadFn != nil {
		return s.getDownloadFn(ctx, reportID, requesterID, role)
	}
	return nil, nil
}

func (s stubReportService) GetStatus(context.Context, int64, int64, string) (*assessmentcommands.ReportExportData, error) {
	return nil, nil
}

func TestDownloadReportStreamsSharedStoreReader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReportHandler(stubReportService{
		getDownloadFn: func(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error) {
			if reportID != 7 || requesterID != 1001 || role != "student" {
				t.Fatalf("unexpected download args: reportID=%d requesterID=%d role=%s", reportID, requesterID, role)
			}
			return &assessmentdomain.ReportDownload{
				Reader:      io.NopCloser(stringsReader("report-binary")),
				Size:        int64(len("report-binary")),
				FileName:    "class-report-7.pdf",
				ContentType: "application/pdf",
			}, nil
		},
	})

	router := gin.New()
	router.GET("/reports/:id/download", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001, Role: "student"})
		handler.DownloadReport(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reports/7/download", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("content-type = %q", got)
	}
	if got := resp.Header().Get("Content-Disposition"); got == "" {
		t.Fatal("expected content-disposition header")
	}
	if resp.Body.String() != "report-binary" {
		t.Fatalf("body = %q", resp.Body.String())
	}
}

func TestDownloadReportPropagatesServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReportHandler(stubReportService{
		getDownloadFn: func(ctx context.Context, reportID, requesterID int64, role string) (*assessmentdomain.ReportDownload, error) {
			return nil, apperror.ErrNotFound
		},
	})

	router := gin.New()
	router.GET("/reports/:id/download", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001, Role: "student"})
		handler.DownloadReport(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/reports/7/download", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", resp.Code, resp.Body.String())
	}
}

type staticReader struct {
	remaining []byte
}

func stringsReader(value string) *staticReader {
	return &staticReader{remaining: []byte(value)}
}

func (r *staticReader) Read(p []byte) (int, error) {
	if len(r.remaining) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.remaining)
	r.remaining = r.remaining[n:]
	return n, nil
}
