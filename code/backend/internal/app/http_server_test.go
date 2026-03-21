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

func TestNewHTTPServerBuildsAndShutsDown(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)
	cfg.Contest.StatusUpdateInterval = time.Second
	cfg.Contest.StatusUpdateBatchSize = 10
	cfg.Contest.StatusUpdateLockTTL = time.Second
	cfg.Contest.AWD.SchedulerInterval = time.Second
	cfg.Contest.AWD.SchedulerLockTTL = time.Second
	cfg.Contest.AWD.SchedulerBatchSize = 10
	cfg.Contest.AWD.RoundInterval = time.Second
	cfg.Contest.AWD.RoundLockTTL = time.Second
	cfg.Contest.AWD.CheckerTimeout = time.Second

	server, err := NewHTTPServer(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewHTTPServer() error = %v", err)
	}
	if server.cleaner == nil || server.assessment == nil || server.statusUpdater == nil || server.awdUpdater == nil {
		t.Fatal("expected http server background components to be initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
