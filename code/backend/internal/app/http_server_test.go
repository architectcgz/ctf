package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"ctf-platform/internal/app/composition"
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
		logger: zap.NewNop(),
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

func TestHTTPServerStartsAndStopsRegisteredBackgroundJobs(t *testing.T) {
	t.Parallel()

	started := make(chan struct{}, 1)
	stopped := make(chan struct{}, 1)
	server := &HTTPServer{
		server: &http.Server{},
		backgroundJobs: []composition.BackgroundJob{
			composition.NewBackgroundJob(
				"test_background_job",
				func(context.Context) error {
					started <- struct{}{}
					return nil
				},
				func(context.Context) error {
					stopped <- struct{}{}
					return nil
				},
			),
		},
		logger: zap.NewNop(),
	}

	if err := server.startBackgroundJobs(); err != nil {
		t.Fatalf("startBackgroundJobs() error = %v", err)
	}

	select {
	case <-started:
	default:
		t.Fatal("expected background job to be started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case <-stopped:
	default:
		t.Fatal("expected background job to be stopped")
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
	if len(server.backgroundJobs) == 0 {
		t.Fatal("expected http server background components to be initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
