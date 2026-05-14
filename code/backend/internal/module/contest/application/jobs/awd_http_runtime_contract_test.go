package jobs

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestAWDRoundUpdaterRunAWDHTTPCheckerActionUsesHTTPRuntime(t *testing.T) {
	runtime := &awdHTTPRuntimeStub{
		response: contestports.AWDHTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "awd{flag-1}",
		},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, "", nil, nil)
	updater.SetHTTPRuntime(runtime)

	result := updater.runAWDHTTPCheckerAction(
		context.Background(),
		"http://service.local:8080",
		"",
		awdHTTPCheckerActionConfig{
			Method:         http.MethodPut,
			Path:           "/api/flag",
			Headers:        map[string]string{"X-Flag": "{{FLAG}}"},
			BodyTemplate:   "{{FLAG}}",
			ExpectedStatus: http.StatusOK,
		},
		awdHTTPCheckerTemplateData{Flag: "awd{flag-1}"},
		[]string{"awd{flag-1}"},
	)

	if result.summary == nil || !result.summary.Healthy {
		t.Fatalf("expected healthy result, got %+v", result.summary)
	}
	if len(runtime.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(runtime.requests))
	}
	request := runtime.requests[0]
	if request.URL != "http://service.local:8080/api/flag" {
		t.Fatalf("unexpected url: %s", request.URL)
	}
	if request.Method != http.MethodPut {
		t.Fatalf("unexpected method: %s", request.Method)
	}
	if request.Headers["X-Flag"] != "awd{flag-1}" {
		t.Fatalf("unexpected headers: %+v", request.Headers)
	}
	if request.Body != "awd{flag-1}" {
		t.Fatalf("unexpected body: %q", request.Body)
	}
	if !request.ReadBody {
		t.Fatal("expected checker action to request response body")
	}
	if request.Timeout != time.Second {
		t.Fatalf("unexpected timeout: %v", request.Timeout)
	}
}

func TestAWDRoundUpdaterRunAWDHTTPCheckerActionMapsReadError(t *testing.T) {
	runtime := &awdHTTPRuntimeStub{
		err: &contestports.AWDHTTPRuntimeError{
			Kind: contestports.AWDHTTPRuntimeErrorKindResponseRead,
			Err:  errors.New("read failed"),
		},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, "", nil, nil)
	updater.SetHTTPRuntime(runtime)

	result := updater.runAWDHTTPCheckerAction(
		context.Background(),
		"http://service.local:8080",
		"",
		awdHTTPCheckerActionConfig{
			Method:         http.MethodGet,
			Path:           "/api/flag",
			ExpectedStatus: http.StatusOK,
		},
		awdHTTPCheckerTemplateData{},
		nil,
	)

	if result.summary == nil {
		t.Fatal("expected summary")
	}
	if result.summary.ErrorCode != "http_response_read_failed" {
		t.Fatalf("unexpected error code: %+v", result.summary)
	}
	if result.summary.Error != "read failed" {
		t.Fatalf("unexpected error message: %+v", result.summary)
	}
}

func TestAWDRoundUpdaterProbeServiceInstanceUsesHTTPRuntime(t *testing.T) {
	runtime := &awdHTTPRuntimeStub{
		response: contestports.AWDHTTPResponse{StatusCode: http.StatusNoContent},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, "", nil, nil)
	updater.SetHTTPRuntime(runtime)

	result := updater.probeServiceInstance(context.Background(), "http://service.local:8080", "", "/health")

	if !result.healthy || result.probe != "http" {
		t.Fatalf("unexpected probe result: %+v", result)
	}
	if len(runtime.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(runtime.requests))
	}
	request := runtime.requests[0]
	if request.Method != http.MethodGet {
		t.Fatalf("unexpected method: %s", request.Method)
	}
	if request.URL != "http://service.local:8080/health" {
		t.Fatalf("unexpected url: %s", request.URL)
	}
	if request.ReadBody {
		t.Fatal("expected probe to skip response body")
	}
	if request.Timeout != time.Second {
		t.Fatalf("unexpected timeout: %v", request.Timeout)
	}
}

type awdHTTPRuntimeStub struct {
	requests  []contestports.AWDHTTPRequest
	response  contestports.AWDHTTPResponse
	responses []contestports.AWDHTTPResponse
	err       error
}

func (s *awdHTTPRuntimeStub) Execute(_ context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
	s.requests = append(s.requests, request)
	if len(s.responses) > 0 {
		response := s.responses[0]
		s.responses = s.responses[1:]
		return response, s.err
	}
	return s.response, s.err
}
