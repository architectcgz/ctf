package commands

import (
	"context"
	"ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	"sync/atomic"
	"testing"
	"time"
)

func TestReportServiceCloseCancelsAsyncTasks(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
			ClassTimeout:  time.Minute,
		},
		nil,
	)
	service.StartBackgroundTasks(context.Background())

	var started atomic.Int32
	startedCh := make(chan struct{})
	service.runAsyncReport(1, func(ctx context.Context) error {
		started.Add(1)
		close(startedCh)
		<-ctx.Done()
		return ctx.Err()
	})

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected async task to start")
	}

	deadlineCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(deadlineCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if started.Load() != 1 {
		t.Fatalf("expected async task to start once, got %d", started.Load())
	}
}

func TestReportServiceCloseRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	if err := service.Close(nil); err == nil {
		t.Fatal("expected Close(nil) to reject missing context")
	}
}

func TestCreatePersonalReportRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := NewReportService(
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		&testReportRepository{},
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)

	_, err := service.CreatePersonalReport(nil, 1, CreatePersonalReportInput{Format: assessmententity.ReportFormatPDF})
	if err == nil {
		t.Fatal("expected CreatePersonalReport(nil) to reject missing context")
	}
}

func TestReportServiceWithPersonalTimeoutUsesConfiguredDeadline(t *testing.T) {
	t.Parallel()

	service := &ReportService{
		config: config.ReportConfig{
			PersonalTimeout: 2 * time.Second,
		},
	}

	ctx, cancel := service.withPersonalTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	remaining := time.Until(deadline)
	if remaining <= time.Second || remaining > 2*time.Second+200*time.Millisecond {
		t.Fatalf("unexpected remaining timeout: %s", remaining)
	}
}
