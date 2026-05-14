package infrastructure

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestAWDHTTPRuntimeAdapterExecuteKeepsAliasHostWhenDialingRuntimeIP(t *testing.T) {
	var seenHost string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenHost = r.Host
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	accessURL := "http://awd-c1-t1-s1:" + parsedURL.Port()
	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-test", Shared: true},
		},
		Containers: []model.InstanceRuntimeContainer{
			{
				ContainerID:    "ctr-alias",
				IsEntryPoint:   true,
				NetworkKeys:    []string{model.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c1-t1-s1"},
				NetworkIPs:     map[string]string{"ctf-awd-test": "127.0.0.1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}

	adapter := NewAWDHTTPRuntimeAdapter(nil, time.Second)
	response, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		AccessURL:      accessURL,
		RuntimeDetails: runtimeDetails,
		URL:            accessURL + "/health",
		Method:         http.MethodGet,
		ReadBody:       true,
		Timeout:        time.Second,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.StatusCode != http.StatusOK || response.Body != "ok" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if !strings.HasPrefix(seenHost, "awd-c1-t1-s1:") {
		t.Fatalf("expected alias host preserved, got %q", seenHost)
	}
}

func TestAWDHTTPRuntimeAdapterExecuteReturnsStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("payload"))
	}))
	t.Cleanup(server.Close)

	adapter := NewAWDHTTPRuntimeAdapter(server.Client(), time.Second)
	response, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		URL:      server.URL + "/body",
		Method:   http.MethodGet,
		ReadBody: true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if response.StatusCode != http.StatusAccepted || response.Body != "payload" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestAWDHTTPRuntimeAdapterExecuteReturnsTypedBuildError(t *testing.T) {
	adapter := NewAWDHTTPRuntimeAdapter(nil, time.Second)
	_, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		URL:      "://bad-url",
		Method:   http.MethodGet,
		ReadBody: true,
	})

	var runtimeErr *contestports.AWDHTTPRuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected typed runtime error, got %T", err)
	}
	if runtimeErr.Kind != contestports.AWDHTTPRuntimeErrorKindRequestBuild {
		t.Fatalf("unexpected error kind: %s", runtimeErr.Kind)
	}
}

func TestAWDHTTPRuntimeAdapterExecuteReturnsTypedReadError(t *testing.T) {
	adapter := NewAWDHTTPRuntimeAdapter(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       errReadCloser{err: errors.New("read failed")},
			}, nil
		}),
	}, time.Second)

	response, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		URL:      "http://example.com",
		Method:   http.MethodGet,
		ReadBody: true,
	})

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response status: %+v", response)
	}
	var runtimeErr *contestports.AWDHTTPRuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected typed runtime error, got %T", err)
	}
	if runtimeErr.Kind != contestports.AWDHTTPRuntimeErrorKindResponseRead {
		t.Fatalf("unexpected error kind: %s", runtimeErr.Kind)
	}
}

func TestAWDHTTPRuntimeAdapterExecuteAppliesDefaultTimeout(t *testing.T) {
	var hasDeadline bool
	var remaining time.Duration
	adapter := NewAWDHTTPRuntimeAdapter(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			deadline, ok := req.Context().Deadline()
			hasDeadline = ok
			if ok {
				remaining = time.Until(deadline)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		}),
	}, 2*time.Second)

	_, err := adapter.Execute(context.Background(), contestports.AWDHTTPRequest{
		URL:      "http://example.com",
		Method:   http.MethodGet,
		ReadBody: false,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !hasDeadline {
		t.Fatal("expected request context deadline")
	}
	if remaining <= 0 || remaining > 2*time.Second {
		t.Fatalf("unexpected remaining timeout: %v", remaining)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type errReadCloser struct {
	err error
}

func (r errReadCloser) Read(_ []byte) (int, error) {
	return 0, r.err
}

func (r errReadCloser) Close() error {
	return nil
}
