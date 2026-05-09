package app

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"ctf-platform/internal/app/composition"
	"go.uber.org/zap"
)

type stubLifecycleCloser struct {
	closed chan struct{}
}

func (s *stubLifecycleCloser) Close(context.Context) error {
	close(s.closed)
	return nil
}

func TestHTTPServerShutdownClosesLifecycleComponents(t *testing.T) {
	t.Parallel()

	reportTasks := &stubLifecycleCloser{closed: make(chan struct{})}
	server := &HTTPServer{
		server: &http.Server{},
		closers: []lifecycleComponent{
			{name: "report_export_tasks", closer: reportTasks},
		},
		logger: zap.NewNop(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case <-reportTasks.closed:
	default:
		t.Fatal("expected lifecycle component to be closed")
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
		appCtx: context.Background(),
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

func TestHTTPServerShutdownStartsHTTPDrainBeforeStoppingBackgroundJobs(t *testing.T) {
	t.Parallel()

	httpShutdownStarted := make(chan struct{})
	httpShutdownRelease := make(chan struct{})
	server := &HTTPServer{
		backgroundJobs: []composition.BackgroundJob{
			composition.NewBackgroundJob(
				"test_background_job",
				nil,
				func(context.Context) error {
					select {
					case <-httpShutdownStarted:
						return nil
					default:
						return errors.New("http shutdown has not started")
					}
				},
			),
		},
		onHTTPShutdownStarted: func() {
			close(httpShutdownStarted)
		},
		shutdownHTTPServer: func(context.Context) error {
			<-httpShutdownRelease
			return nil
		},
		logger: zap.NewNop(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Shutdown(ctx)
	}()

	select {
	case <-httpShutdownStarted:
	case <-ctx.Done():
		t.Fatal("expected http shutdown to start before timeout")
	}

	close(httpShutdownRelease)

	if err := <-done; err != nil {
		t.Fatalf("Shutdown() error = %v", err)
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
