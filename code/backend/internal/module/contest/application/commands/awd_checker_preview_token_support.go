package commands

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

const awdCheckerPreviewTokenTTL = 30 * time.Minute

func storeAWDCheckerPreviewToken(
	ctx context.Context,
	store contestports.AWDCheckerPreviewTokenStore,
	contestID, serviceID, awdChallengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	checkerTokenEnv string,
	result *dto.AWDCheckerPreviewResp,
) (string, error) {
	if result == nil {
		return "", nil
	}
	if store == nil {
		return "", contestports.ErrAWDCheckerPreviewTokenStoreUnavailable
	}

	record := contestports.AWDCheckerPreviewTokenRecord{
		ContestID:       contestID,
		ServiceID:       serviceID,
		AWDChallengeID:  awdChallengeID,
		CheckerType:     checkerType,
		CheckerConfig:   checkerConfig,
		CheckerTokenEnv: strings.TrimSpace(checkerTokenEnv),
		Result:          *awdPreviewResultMapper.ToDomainPtr(result),
		CreatedAt:       time.Now().UTC(),
	}
	return store.StoreAWDCheckerPreviewToken(ctx, record, awdCheckerPreviewTokenTTL)
}

func consumeCheckerPreviewValidationState(
	ctx context.Context,
	store contestports.AWDCheckerPreviewTokenStore,
	contestID, serviceID, awdChallengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	checkerTokenEnv string,
	previewToken string,
) (model.AWDCheckerValidationState, *time.Time, string, error) {
	if strings.TrimSpace(previewToken) != "" && store == nil {
		return model.AWDCheckerValidationStatePending, nil, "", errcode.ErrAWDCheckerPreviewUnavailable
	}
	record, err := consumeAWDCheckerPreviewToken(ctx, store, contestID, serviceID, awdChallengeID, checkerType, checkerConfig, checkerTokenEnv, previewToken)
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
	store contestports.AWDCheckerPreviewTokenStore,
	contestID, serviceID, awdChallengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	checkerTokenEnv string,
	previewToken string,
) (*contestports.AWDCheckerPreviewTokenRecord, error) {
	if store == nil || strings.TrimSpace(previewToken) == "" {
		return nil, nil
	}
	record, found, err := store.LoadAWDCheckerPreviewToken(ctx, contestID, previewToken)
	if err != nil {
		if err == contestports.ErrAWDCheckerPreviewTokenStoreUnavailable {
			return nil, errcode.ErrAWDCheckerPreviewUnavailable
		}
		return nil, err
	}
	if !found || record == nil {
		return nil, nil
	}
	if !matchesAWDCheckerPreviewTokenRecord(record, contestID, serviceID, awdChallengeID, checkerType, checkerConfig, checkerTokenEnv) {
		return nil, nil
	}
	if err := store.DeleteAWDCheckerPreviewToken(ctx, contestID, previewToken); err != nil {
		if err == contestports.ErrAWDCheckerPreviewTokenStoreUnavailable {
			return nil, errcode.ErrAWDCheckerPreviewUnavailable
		}
		return nil, err
	}
	return record, nil
}

func matchesAWDCheckerPreviewTokenRecord(
	record *contestports.AWDCheckerPreviewTokenRecord,
	contestID, serviceID, awdChallengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	checkerTokenEnv string,
) bool {
	if record == nil {
		return false
	}
	if record.ContestID != contestID ||
		record.CheckerType != checkerType ||
		record.CheckerConfig != checkerConfig ||
		strings.TrimSpace(record.CheckerTokenEnv) != strings.TrimSpace(checkerTokenEnv) {
		return false
	}
	if serviceID > 0 || record.ServiceID > 0 {
		return record.ServiceID == serviceID && record.AWDChallengeID == awdChallengeID
	}
	return record.ContestID == contestID &&
		record.ServiceID == serviceID &&
		record.AWDChallengeID == awdChallengeID &&
		record.ServiceID == 0
}
