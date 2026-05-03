package commands

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
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

func (s *Service) triggerAssessmentUpdate(userID int64, dimension string) {
	if s.assessmentService == nil || !model.IsValidDimension(dimension) {
		return
	}

	s.runAsyncTask(func(ctx context.Context) {
		timer := time.NewTimer(s.config.Assessment.IncrementalUpdateDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}

		updateCtx, cancel := context.WithTimeout(ctx, s.config.Assessment.IncrementalUpdateTimeout)
		defer cancel()

		if err := s.assessmentService.UpdateSkillProfileForDimension(updateCtx, userID, dimension); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("更新能力画像失败",
				zap.Int64("user_id", userID),
				zap.String("dimension", dimension),
				zap.Error(err))
		}
	})
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
