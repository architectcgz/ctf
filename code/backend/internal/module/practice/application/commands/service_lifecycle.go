package commands

import (
	"context"
	"errors"

	"go.uber.org/zap"

	platformevents "ctf-platform/internal/platform/events"
	"ctf-platform/internal/shared/safego"
)

func (s *serviceCore) StartBackgroundTasks(ctx context.Context) {
	if s == nil || ctx == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.baseCtx, s.cancel = context.WithCancel(ctx)
}

func (s *serviceCore) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("practice service close requires context")
	}
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	safego.Go(nil, ctx, s.logger, "practice_service_close_wait", func(context.Context) {
		s.tasks.Wait()
		close(done)
	})

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *serviceCore) triggerScoreUpdate(userID int64) {
	if s.scoreService == nil {
		return
	}

	s.runAsyncTask(func(ctx context.Context) {
		scoreCtx := ctx
		cancel := func() {}
		if timeout := s.scoreService.lockTimeout(); timeout > 0 {
			scoreCtx, cancel = context.WithTimeout(ctx, timeout)
		}
		defer cancel()

		if err := s.scoreService.UpdateUserScore(scoreCtx, userID); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("更新用户得分失败", zap.Int64("user_id", userID), zap.Error(err))
		}
	})
}

func (s *serviceCore) runAsyncTask(fn func(context.Context)) {
	if s.baseCtx == nil || fn == nil {
		return
	}

	safego.Go(&s.tasks, s.baseCtx, s.logger, "practice_async_task", func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fn(ctx)
	})
}

func (s *serviceCore) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s.eventBus == nil {
		return
	}
	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.logger.Warn("publish_practice_event_failed", zap.String("event", evt.Name), zap.Error(err))
	}
}
