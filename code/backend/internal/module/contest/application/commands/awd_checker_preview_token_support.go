package commands

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

const awdCheckerPreviewTokenTTL = 30 * time.Minute

type storedAWDCheckerPreviewToken struct {
	ContestID     int64                     `json:"contest_id"`
	ServiceID     int64                     `json:"service_id"`
	ChallengeID   int64                     `json:"challenge_id"`
	CheckerType   model.AWDCheckerType      `json:"checker_type"`
	CheckerConfig string                    `json:"checker_config"`
	Result        dto.AWDCheckerPreviewResp `json:"result"`
	CreatedAt     time.Time                 `json:"created_at"`
}

func storeAWDCheckerPreviewToken(
	ctx context.Context,
	redisClient *redislib.Client,
	contestID, serviceID, challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	result *dto.AWDCheckerPreviewResp,
) (string, error) {
	if redisClient == nil || result == nil {
		return "", nil
	}

	token := uuid.NewString()
	record := storedAWDCheckerPreviewToken{
		ContestID:     contestID,
		ServiceID:     serviceID,
		ChallengeID:   challengeID,
		CheckerType:   checkerType,
		CheckerConfig: checkerConfig,
		Result:        *result,
		CreatedAt:     time.Now().UTC(),
	}
	raw, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	if err := redisClient.Set(ctx, rediskeys.AWDCheckerPreviewTokenKey(contestID, token), raw, awdCheckerPreviewTokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func consumeCheckerPreviewValidationState(
	ctx context.Context,
	redisClient *redislib.Client,
	contestID, serviceID, challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	previewToken string,
) (model.AWDCheckerValidationState, *time.Time, string, error) {
	if strings.TrimSpace(previewToken) != "" && redisClient == nil {
		return model.AWDCheckerValidationStatePending, nil, "", errcode.ErrAWDCheckerPreviewUnavailable
	}
	record, err := consumeAWDCheckerPreviewToken(ctx, redisClient, contestID, serviceID, challengeID, checkerType, checkerConfig, previewToken)
	if err != nil {
		return model.AWDCheckerValidationStatePending, nil, "", err
	}
	if record == nil {
		return model.AWDCheckerValidationStatePending, nil, "", nil
	}

	checkedAt := record.CreatedAt
	if value, ok := record.Result.CheckResult["checked_at"].(string); ok {
		if parsed, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(value)); parseErr == nil {
			checkedAt = parsed
		}
	}
	rawResult, err := contestdomain.MarshalAWDCheckerPreviewResult(&record.Result)
	if err != nil {
		return model.AWDCheckerValidationStatePending, nil, "", err
	}

	state := model.AWDCheckerValidationStateFailed
	if record.Result.ServiceStatus == model.AWDServiceStatusUp {
		state = model.AWDCheckerValidationStatePassed
	}
	return state, &checkedAt, rawResult, nil
}

func consumeAWDCheckerPreviewToken(
	ctx context.Context,
	redisClient *redislib.Client,
	contestID, serviceID, challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	previewToken string,
) (*storedAWDCheckerPreviewToken, error) {
	if redisClient == nil || strings.TrimSpace(previewToken) == "" {
		return nil, nil
	}

	key := rediskeys.AWDCheckerPreviewTokenKey(contestID, previewToken)
	raw, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, nil
		}
		return nil, err
	}

	var record storedAWDCheckerPreviewToken
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		return nil, err
	}
	if !record.matches(contestID, serviceID, challengeID, checkerType, checkerConfig) {
		return nil, nil
	}
	if err := redisClient.Del(ctx, key).Err(); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r storedAWDCheckerPreviewToken) matches(
	contestID, serviceID, challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
) bool {
	if r.ContestID != contestID ||
		r.CheckerType != checkerType ||
		r.CheckerConfig != checkerConfig {
		return false
	}
	if serviceID > 0 || r.ServiceID > 0 {
		return r.ServiceID == serviceID && r.ChallengeID == challengeID
	}
	return r.ContestID == contestID &&
		r.ServiceID == serviceID &&
		r.ChallengeID == challengeID &&
		r.ServiceID == 0
}
