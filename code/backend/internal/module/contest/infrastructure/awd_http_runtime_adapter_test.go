package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

type awdHTTPRuntimeRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f awdHTTPRuntimeRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type awdHTTPRuntimeFailingBody struct{}

func (awdHTTPRuntimeFailingBody) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (awdHTTPRuntimeFailingBody) Close() error {
	return nil
}

func TestAWDHTTPRuntimeAdapterDialsRuntimeIPButPreservesAliasHost(t *testing.T) {
	t.Parallel()

	hostCh := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostCh <- r.Host
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	aliasURL := fmt.Sprintf("http://awd-c8-t15-s21:%s/health", serverURL.Port())
	adapter := NewAWDHTTPRuntimeAdapter(server.Client(), time.Second)

	response, execErr := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL:      fmt.Sprintf("http://awd-c8-t15-s21:%s", serverURL.Port()),
		RuntimeDetails: awdHTTPRuntimeDetailsJSON(serverURL.Hostname()),
		URL:            aliasURL,
		Method:         http.MethodGet,
	})
	if execErr != nil {
		t.Fatalf("execute: %v", execErr)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}

	select {
	case host := <-hostCh:
		if host != fmt.Sprintf("awd-c8-t15-s21:%s", serverURL.Port()) {
			t.Fatalf("host = %q, want %q", host, fmt.Sprintf("awd-c8-t15-s21:%s", serverURL.Port()))
		}
	case <-time.After(time.Second):
		t.Fatal("server did not receive request")
	}
}

func TestAWDHTTPRuntimeAdapterReturnsBodyWhenRequested(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "created")
	}))
	t.Cleanup(server.Close)

	adapter := NewAWDHTTPRuntimeAdapter(server.Client(), time.Second)
	response, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL: server.URL,
		URL:       server.URL,
		Method:    http.MethodPost,
		Body:      "payload",
		ReadBody:  true,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	if response.Body != "created" {
		t.Fatalf("body = %q, want %q", response.Body, "created")
	}
}

func TestAWDHTTPRuntimeAdapterReturnsTypedRequestBuildError(t *testing.T) {
	t.Parallel()

	adapter := NewAWDHTTPRuntimeAdapter(&http.Client{}, time.Second)
	_, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL: "http://awd-c8-t15-s21:8080",
		URL:       "http://[::1",
		Method:    http.MethodGet,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var runtimeErr *contestports.AWDHTTPRuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("error type = %T, want *AWDHTTPRuntimeError", err)
	}
	if runtimeErr.Kind != contestports.AWDHTTPRuntimeErrorKindRequestBuild {
		t.Fatalf("kind = %q, want %q", runtimeErr.Kind, contestports.AWDHTTPRuntimeErrorKindRequestBuild)
	}
	if runtimeErr.Unwrap() == nil {
		t.Fatal("expected wrapped error")
	}
}

func TestAWDHTTPRuntimeAdapterReturnsTypedResponseReadError(t *testing.T) {
	t.Parallel()

	adapter := NewAWDHTTPRuntimeAdapter(&http.Client{
		Transport: awdHTTPRuntimeRoundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       awdHTTPRuntimeFailingBody{},
				Header:     make(http.Header),
			}, nil
		}),
	}, time.Second)

	_, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL: "http://service.local",
		URL:       "http://service.local/health",
		Method:    http.MethodGet,
		ReadBody:  true,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var runtimeErr *contestports.AWDHTTPRuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("error type = %T, want *AWDHTTPRuntimeError", err)
	}
	if runtimeErr.Kind != contestports.AWDHTTPRuntimeErrorKindResponseRead {
		t.Fatalf("kind = %q, want %q", runtimeErr.Kind, contestports.AWDHTTPRuntimeErrorKindResponseRead)
	}
}

func TestAWDHTTPRuntimeAdapterAppliesDefaultTimeoutWhenRequestTimeoutMissing(t *testing.T) {
	t.Parallel()

	checked := false
	adapter := NewAWDHTTPRuntimeAdapter(&http.Client{
		Transport: awdHTTPRuntimeRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			if !ok {
				t.Fatal("expected request deadline")
			}
			remaining := time.Until(deadline)
			if remaining < 2*time.Second || remaining > 4*time.Second {
				t.Fatalf("remaining timeout = %s, want between 2s and 4s", remaining)
			}
			checked = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}, 0)

	response, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL: "http://service.local",
		URL:       "http://service.local/health",
		Method:    http.MethodGet,
		ReadBody:  true,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !checked {
		t.Fatal("expected timeout check to run")
	}
	if response.Body != "ok" {
		t.Fatalf("body = %q, want %q", response.Body, "ok")
	}
}

func awdHTTPRuntimeDetailsJSON(ip string) string {
	return fmt.Sprintf(`{"networks":[{"key":"default","name":"ctf-awd-contest-8"}],"containers":[{"is_entry_point":true,"network_keys":["default"],"network_aliases":["awd-c8-t15-s21"],"network_ips":{"ctf-awd-contest-8":"%s"}}]}`, ip)
}
