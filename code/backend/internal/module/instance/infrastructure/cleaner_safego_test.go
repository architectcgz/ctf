package infrastructure

import (
	"context"
	"testing"
	"time"
)

func TestCleanerStartRunOnceRequiresSafeGoRecovery(t *testing.T) {
	t.Parallel()

	cleaner := &Cleaner{
		baseCtx: context.Background(),
		cancel:  func() {},
	}

	cleaner.startRunOnce()
	cleaner.wg.Wait()
	_ = time.Second
}
