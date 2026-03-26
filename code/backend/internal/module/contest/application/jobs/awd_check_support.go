package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) shouldSyncLiveServiceStatusCache(ctx context.Context, contestID int64, round *model.AWDRound) (bool, error) {
	if u.redis == nil || u.repo == nil || contestID <= 0 || round == nil {
		return false, nil
	}

	currentRound, err := u.repo.FindRunningRound(ctx, contestID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
		currentRoundNumber, redisErr := u.redis.Get(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Result()
		if redisErr == nil {
			return strings.TrimSpace(currentRoundNumber) == fmt.Sprintf("%d", round.RoundNumber), nil
		}
		if !errors.Is(redisErr, redislib.Nil) {
			return false, redisErr
		}
		return false, nil
	}
	return currentRound.ID == round.ID, nil
}

func (u *AWDRoundUpdater) loadContestServiceInstances(ctx context.Context, contestID int64, challenges []model.Challenge) ([]contestports.AWDServiceInstance, error) {
	if len(challenges) == 0 {
		return nil, nil
	}

	challengeIDs := make([]int64, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}

	return u.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
}
