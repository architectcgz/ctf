package safego

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"
)

type inertContext struct{}

func (inertContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (inertContext) Done() <-chan struct{} {
	return nil
}

func (inertContext) Err() error {
	return nil
}

func (inertContext) Value(any) any {
	return nil
}

var fallbackContext context.Context = inertContext{}

func Go(wg *sync.WaitGroup, ctx context.Context, logger *zap.Logger, taskName string, fn func(context.Context)) {
	if fn == nil {
		return
	}
	if ctx == nil {
		ctx = fallbackContext
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic_recovered",
					zap.Any("panic", recovered),
					zap.String("task_name", taskName),
					zap.ByteString("stack", debug.Stack()),
				)
			}
		}()
		fn(ctx)
	}()
}

func GoWait(ctx context.Context, logger *zap.Logger, taskName string, fn func(context.Context)) {
	Go(nil, ctx, logger, taskName, fn)
}

func GoWithRecover(ctx context.Context, logger *zap.Logger, taskName string, fn func(context.Context), wg *sync.WaitGroup) {
	Go(wg, ctx, logger, taskName, fn)
}
