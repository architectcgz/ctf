package composition

import (
	"context"
	"errors"
	"sync"
)

func NewLoopBackgroundJob(name string, run func(context.Context)) BackgroundJob {
	var (
		mu      sync.Mutex
		cancel  context.CancelFunc
		started bool
		wg      sync.WaitGroup
	)

	return NewBackgroundJob(
		name,
		func(ctx context.Context) error {
			if ctx == nil {
				return errors.New("background job start requires context")
			}
			mu.Lock()
			defer mu.Unlock()
			if started {
				return nil
			}
			started = true

			runCtx, runCancel := context.WithCancel(ctx)
			cancel = runCancel
			wg.Add(1)
			go func() {
				defer wg.Done()
				run(runCtx)
			}()
			return nil
		},
		func(ctx context.Context) error {
			mu.Lock()
			if !started {
				mu.Unlock()
				return nil
			}
			stopFn := cancel
			mu.Unlock()

			if stopFn != nil {
				stopFn()
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	)
}
