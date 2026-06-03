package fullrouterreportstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
)

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type ReportRecord struct {
	ID int64
}

type ReportPreviewAndDownloadStateMatrixDriver struct {
	Request                RequestFunc
	AdminHeaders           map[string]string
	TeacherHeaders         map[string]string
	StudentHeaders         map[string]string
	ClassName              string
	OtherClassName         string
	ContestID              int64
	StudentID              int64
	OtherStudentID         int64
	CreateProcessingReport func(t *testing.T) ReportRecord
	CreateFailedReport     func(t *testing.T, message string) ReportRecord
	WaitForReportStatus    func(t *testing.T, reportID int64, headers map[string]string, wantStatus string) *assessmentcmd.ReportExportData
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyReportPreviewAndDownloadStateMatrix(t *testing.T, driver ReportPreviewAndDownloadStateMatrixDriver) {
	t.Helper()

	resp := driver.Request(http.MethodPost, "/api/v1/reports/personal", map[string]any{
		"format": assessmententity.ReportFormatExcel,
	}, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var personalReport assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &personalReport)
	if personalReport.Status != assessmententity.ReportStatusReady || personalReport.DownloadURL == nil {
		t.Fatalf("expected ready personal report with download url, got %+v", personalReport)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", personalReport.ReportID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var personalStatus assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &personalStatus)
	if personalStatus.Status != assessmententity.ReportStatusReady || personalStatus.DownloadURL == nil {
		t.Fatalf("expected ready personal report status, got %+v", personalStatus)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", personalReport.ReportID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("expected xlsx content-type, got %q", contentType)
	}

	processingReport := driver.CreateProcessingReport(t)
	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", processingReport.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var processingStatus assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &processingStatus)
	if processingStatus.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", processingStatus)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", processingReport.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusConflict)

	failedReport := driver.CreateFailedReport(t, "generation failed in matrix")
	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", failedReport.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusOK)

	var failedStatus assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &failedStatus)
	if failedStatus.Status != assessmententity.ReportStatusFailed || failedStatus.ErrorMessage == nil || *failedStatus.ErrorMessage != "generation failed in matrix" {
		t.Fatalf("expected failed status with message, got %+v", failedStatus)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", failedReport.ID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusConflict)

	resp = driver.Request(http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": driver.OtherClassName,
		"format":     assessmententity.ReportFormatPDF,
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, "/api/v1/reports/class", map[string]any{
		"class_name": driver.ClassName,
		"format":     assessmententity.ReportFormatPDF,
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var classReport assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &classReport)
	if classReport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected class report to start in processing state, got %+v", classReport)
	}

	classReady := driver.WaitForReportStatus(t, classReport.ReportID, driver.TeacherHeaders, assessmententity.ReportStatusReady)
	if classReady == nil || classReady.DownloadURL == nil {
		t.Fatalf("expected class report download url after ready, got %+v", classReady)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", classReport.ReportID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/pdf" {
		t.Fatalf("expected pdf content-type, got %q", contentType)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", classReport.ReportID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", classReport.ReportID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/admin/contests/%d/export", driver.ContestID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)

	var contestExport assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &contestExport)
	if contestExport.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected contest export processing status, got %+v", contestExport)
	}

	contestExportReady := driver.WaitForReportStatus(t, contestExport.ReportID, driver.AdminHeaders, assessmententity.ReportStatusReady)
	if contestExportReady == nil || contestExportReady.DownloadURL == nil {
		t.Fatalf("expected contest export download url, got %+v", contestExportReady)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", contestExport.ReportID), nil, driver.AdminHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type, got %q", contentType)
	}
	if !strings.Contains(resp.Body.String(), "\"contest\"") {
		t.Fatalf("expected contest export json payload, got %s", resp.Body.String())
	}

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", driver.OtherStudentID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusForbidden)

	resp = driver.Request(http.MethodPost, fmt.Sprintf("/api/v1/teacher/students/%d/review-archive/export", driver.StudentID), map[string]any{
		"format": assessmententity.ReportFormatJSON,
	}, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)

	var reviewArchive assessmentcmd.ReportExportData
	decodeEnvelopeData(t, resp, &reviewArchive)
	if reviewArchive.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected review archive processing status, got %+v", reviewArchive)
	}

	reviewArchiveReady := driver.WaitForReportStatus(t, reviewArchive.ReportID, driver.TeacherHeaders, assessmententity.ReportStatusReady)
	if reviewArchiveReady == nil || reviewArchiveReady.DownloadURL == nil {
		t.Fatalf("expected review archive download url, got %+v", reviewArchiveReady)
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/download", reviewArchive.ReportID), nil, driver.TeacherHeaders)
	assertStatus(t, resp, http.StatusOK)
	if contentType := resp.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected json content-type, got %q", contentType)
	}
	if !strings.Contains(resp.Body.String(), "\"manual_reviews\"") {
		t.Fatalf("expected review archive json payload, got %s", resp.Body.String())
	}

	resp = driver.Request(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d", reviewArchive.ReportID), nil, driver.StudentHeaders)
	assertStatus(t, resp, http.StatusForbidden)
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("expected status %d, got %d body=%s", want, resp.Code, resp.Body.String())
	}
}

func decodeEnvelopeData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}
