package commands

import (
	"context"
	"testing"
	"time"
)

func TestRunAsyncTaskRequiresSafeGoForPanicRecovery(t *testing.T) {
	t.Parallel()

	service := newServiceCore(nil, nil, nil, nil, nil, nil, nil, nil)
	service.StartBackgroundTasks(context.Background())

	started := make(chan struct{})
	service.runAsyncTask(func(context.Context) {
		close(started)
		panic("boom")
	})

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected async task to start")
	}

	if err := service.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
