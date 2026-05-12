package commands

import (
	"context"
	"errors"

	"go.uber.org/zap"

	platformevents "ctf-platform/internal/platform/events"
)

func (s *Service) StartBackgroundTasks(ctx context.Context) {
	if s == nil || ctx == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.baseCtx, s.cancel = context.WithCancel(ctx)
}

func (s *Service) Close(ctx context.Context) error {
	if ctx == nil {
		return errors.New("practice service close requires context")
	}
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	go func() {
		s.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) triggerScoreUpdate(userID int64) {
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

func (s *Service) runAsyncTask(fn func(context.Context)) {
	if s.baseCtx == nil {
		return
	}

	s.tasks.Add(1)
	go func() {
		defer s.tasks.Done()

		select {
		case <-s.baseCtx.Done():
			return
		default:
		}

		fn(s.baseCtx)
	}()
}

func (s *Service) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s.eventBus == nil {
		return
	}
	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.logger.Warn("publish_practice_event_failed", zap.String("event", evt.Name), zap.Error(err))
	}
}
