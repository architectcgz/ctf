package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/internal/platform/events"
	"gorm.io/gorm"
	"sync/atomic"
	"testing"
	"time"
)

func TestPracticeServiceCloseCancelsAsyncScoreUpdate(t *testing.T) {
	t.Parallel()

	startedCh := make(chan struct{})
	var calls atomic.Int32
	service := NewService(
		&stubPracticeRepository{
			findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
				return nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		&stubScoreUpdater{
			lockWait: time.Minute,
			updateFn: func(ctx context.Context, userID int64) error {
				if userID != 7 {
					t.Fatalf("unexpected userID: %d", userID)
				}
				calls.Add(1)
				close(startedCh)
				<-ctx.Done()
				return ctx.Err()
			},
		},
		nil,
		&config.Config{},
		nil)

	service.StartBackgroundTasks(context.Background())

	service.triggerScoreUpdate(7)

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected score update to start")
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one score update call, got %d", calls.Load())
	}
}

func TestMarkInstanceFailedDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	service := NewService(
		nil,
		nil,
		nil,
		&stubPracticeInstanceStore{
			updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
				if ctx != nil {
					t.Fatalf("expected update status ctx to stay nil, got %v", ctx)
				}
				return nil
			},
		},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *model.Instance) error {
				if ctx != nil {
					t.Fatalf("expected cleanup ctx to stay nil, got %v", ctx)
				}
				return nil
			},
		},
		nil,
		nil,
		nil,
		nil)

	service.markInstanceFailed(nil, &model.Instance{ID: 42})
}

func TestPublishWeakEventDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	publishCalled := false
	service := NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil).
		SetEventBus(&stubPracticeEventBus{
			publishFn: func(ctx context.Context, evt events.Event) error {
				publishCalled = true
				if ctx != nil {
					t.Fatalf("expected publish ctx to stay nil, got %v", ctx)
				}
				return nil
			},
		})

	service.publishWeakEvent(nil, events.Event{Name: "practice.test"})
	if !publishCalled {
		t.Fatal("expected event to be published")
	}
}

func TestPracticeServiceRunAsyncTaskReturnsWhenClosed(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)
	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("initial Close() error = %v", err)
	}

	called := make(chan struct{}, 1)
	service.runAsyncTask(func(context.Context) {
		called <- struct{}{}
	})

	select {
	case <-called:
		t.Fatal("expected closed service to skip async task")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRunProvisioningLoopReturnsWhenContextMissing(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, nil, nil, nil, nil, nil, &config.Config{
		Container: config.ContainerConfig{
			Scheduler: config.ContainerSchedulerConfig{
				Enabled: true,
			},
		},
	}, nil)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("RunProvisioningLoop(nil) should return without panic, got %v", recovered)
		}
	}()
	service.RunProvisioningLoop(nil)
}

func TestPracticeServiceCloseRejectsNilContext(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	if err := service.Close(nil); err == nil {
		t.Fatal("expected Close(nil) to reject missing context")
	}
}
