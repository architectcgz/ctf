package infrastructure

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInstanceReadinessProbeAcceptsHTTPAccessURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	probe := NewInstanceReadinessProbe()
	if err := probe.ProbeAccessURL(context.Background(), server.URL, time.Second); err != nil {
		t.Fatalf("ProbeAccessURL() error = %v", err)
	}
}

func TestInstanceReadinessProbeAcceptsTCPAccessURL(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{}, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		_ = conn.Close()
		accepted <- struct{}{}
	}()

	probe := NewInstanceReadinessProbe()
	if err := probe.ProbeAccessURL(context.Background(), "tcp://"+listener.Addr().String(), time.Second); err != nil {
		t.Fatalf("ProbeAccessURL() error = %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected tcp probe to connect")
	}
}

func TestInstanceReadinessProbeRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	probe := NewInstanceReadinessProbe()
	if err := probe.ProbeAccessURL(context.Background(), "://bad-url", time.Second); err == nil {
		t.Fatal("expected invalid url error")
	}
}
