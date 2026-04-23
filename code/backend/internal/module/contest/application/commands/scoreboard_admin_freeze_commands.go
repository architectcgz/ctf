package commands

import (
	"context"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
	ctfws "ctf-platform/pkg/websocket"
)

func (s *ScoreboardAdminService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
		return errcode.ErrContestEnded
	}

	freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime
	snapshotCreated := false
	if !now.Before(freezeTime) {
		contest.Status = model.ContestStatusFrozen
		if err := s.createSnapshotFromLive(ctx, contestID); err != nil {
			return err
		}
		snapshotCreated = true
	}

	if err := s.repo.Update(ctx, contest); err != nil {
		if snapshotCreated {
			if rollbackErr := s.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err(); rollbackErr != nil {
				return errcode.ErrInternal.WithCause(fmt.Errorf("update contest freeze state: %w; rollback frozen snapshot: %v", err, rollbackErr))
			}
		}
		return err
	}

	broadcastContestRealtimeEvent(s.broadcaster, contestports.ScoreboardChannel(contestID), ctfws.Envelope{
		Type: "scoreboard.updated",
		Payload: map[string]any{
			"contest_id": contestID,
		},
	})
	return nil
}

func (s *ScoreboardAdminService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != model.ContestStatusFrozen {
		return errcode.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	if contest.Status == model.ContestStatusFrozen && time.Now().Before(contest.EndTime) {
		contest.Status = model.ContestStatusRunning
	}
	if err := s.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	if err := s.repo.Update(ctx, contest); err != nil {
		return err
	}

	broadcastContestRealtimeEvent(s.broadcaster, contestports.ScoreboardChannel(contestID), ctfws.Envelope{
		Type: "scoreboard.updated",
		Payload: map[string]any{
			"contest_id": contestID,
		},
	})
	return nil
}

func (s *ScoreboardAdminService) createSnapshotFromLive(ctx context.Context, contestID int64) error {
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
