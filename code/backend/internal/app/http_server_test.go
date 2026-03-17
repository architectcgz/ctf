package app

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

type stubReportServiceCloser struct {
	closed chan struct{}
}

func (s *stubReportServiceCloser) Close(context.Context) error {
	close(s.closed)
	return nil
}

func TestHTTPServerShutdownClosesReportService(t *testing.T) {
	t.Parallel()

	reportCloser := &stubReportServiceCloser{closed: make(chan struct{})}
	server := &HTTPServer{
		server: &http.Server{},
		closers: []lifecycleComponent{
			{name: "report_service", closer: reportCloser},
		},
		updaterWG: &sync.WaitGroup{},
		logger:    zap.NewNop(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case <-reportCloser.closed:
	default:
		t.Fatal("expected report service to be closed")
	}
}
