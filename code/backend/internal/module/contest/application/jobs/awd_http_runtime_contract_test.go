package jobs

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

type fakeAWDHTTPRuntime struct {
	requests []contestports.AWDHTTPRequest
	execute  func(context.Context, contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error)
}

func (f *fakeAWDHTTPRuntime) Execute(ctx context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
	f.requests = append(f.requests, request)
	if f.execute != nil {
		return f.execute(ctx, request)
	}
	return contestports.AWDHTTPResponse{}, nil
}

func TestAWDHTTPRuntimeContractCheckerActionUsesPortRequestAndBody(t *testing.T) {
	runtime := &fakeAWDHTTPRuntime{
		execute: func(_ context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
			if request.AccessURL != "http://service.local/base" {
				t.Fatalf("unexpected access url: %q", request.AccessURL)
			}
			if request.RuntimeDetails != "runtime-details" {
				t.Fatalf("unexpected runtime details: %q", request.RuntimeDetails)
			}
			if request.URL != "http://service.local/base/api/flag" {
				t.Fatalf("unexpected target url: %q", request.URL)
			}
			if request.Method != "PUT" {
				t.Fatalf("unexpected method: %q", request.Method)
			}
			if request.Headers["X-Flag"] != "awd{team-1}" {
				t.Fatalf("unexpected rendered header: %#v", request.Headers)
			}
			if request.Body != "awd{team-1}" {
				t.Fatalf("unexpected body: %q", request.Body)
			}
			if !request.ReadBody {
				t.Fatal("expected checker action to read response body")
			}
			if request.Timeout != 2*time.Second {
				t.Fatalf("unexpected timeout: %s", request.Timeout)
			}
			return contestports.AWDHTTPResponse{StatusCode: 200, Body: "awd{team-1}"}, nil
		},
	}
	updater := &AWDRoundUpdater{
		cfg:         config.ContestAWDConfig{CheckerTimeout: 2 * time.Second},
		httpRuntime: runtime,
	}

	result := updater.runAWDHTTPCheckerAction(
		context.Background(),
		"http://service.local/base",
		"runtime-details",
		awdHTTPCheckerActionConfig{
			Method:         "PUT",
			Path:           "/api/flag",
			Headers:        map[string]string{"X-Flag": "{{FLAG}}"},
			BodyTemplate:   "{{FLAG}}",
			ExpectedStatus: 200,
		},
		awdHTTPCheckerTemplateData{Flag: "awd{team-1}"},
		[]string{"awd{team-1}"},
	)

	if len(runtime.requests) != 1 {
		t.Fatalf("expected 1 runtime request, got %d", len(runtime.requests))
	}
	if !result.summary.Healthy || result.responseBody != "awd{team-1}" {
		t.Fatalf("unexpected checker result: %+v", result)
	}
}

func TestAWDHTTPRuntimeContractProbeUsesGETWithoutBodyRead(t *testing.T) {
	runtime := &fakeAWDHTTPRuntime{
		execute: func(_ context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
			if request.Method != "GET" {
				t.Fatalf("unexpected method: %q", request.Method)
			}
			if request.URL != "http://service.local/health" {
				t.Fatalf("unexpected probe url: %q", request.URL)
			}
			if request.ReadBody {
				t.Fatal("expected probe to skip response body reads")
			}
			if request.Body != "" {
				t.Fatalf("unexpected probe body: %q", request.Body)
			}
			return contestports.AWDHTTPResponse{StatusCode: 204}, nil
		},
	}
	updater := &AWDRoundUpdater{
		cfg:         config.ContestAWDConfig{CheckerTimeout: 1500 * time.Millisecond},
		httpRuntime: runtime,
	}

	result := updater.probeServiceInstance(context.Background(), "http://service.local", "runtime-details", "/health")

	if len(runtime.requests) != 1 {
		t.Fatalf("expected 1 runtime request, got %d", len(runtime.requests))
	}
	if !result.healthy || result.probe != "http" {
		t.Fatalf("unexpected probe result: %+v", result)
	}
}

func TestAWDHTTPRuntimeContractResponseReadErrorMapsToResponseReadFailure(t *testing.T) {
	runtime := &fakeAWDHTTPRuntime{
		execute: func(_ context.Context, _ contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
			return contestports.AWDHTTPResponse{StatusCode: 200}, &contestports.AWDHTTPRuntimeError{Kind: contestports.AWDHTTPRuntimeErrorKindResponseRead}
		},
	}
	updater := &AWDRoundUpdater{
		cfg:         config.ContestAWDConfig{CheckerTimeout: time.Second},
		httpRuntime: runtime,
	}

	result := updater.runAWDHTTPCheckerAction(
		context.Background(),
		"http://service.local",
		"runtime-details",
		awdHTTPCheckerActionConfig{Method: "GET", Path: "/health", ExpectedStatus: 200},
		awdHTTPCheckerTemplateData{},
		nil,
	)

	if result.summary.ErrorCode != "http_response_read_failed" {
		t.Fatalf("expected response read failure mapping, got %+v", result.summary)
	}
	if result.summary.StatusCode != 200 {
		t.Fatalf("expected status code from runtime error response, got %+v", result.summary)
	}
}
