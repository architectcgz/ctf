package queries

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ContestAWDServiceQueryService struct {
	repo        contestports.AWDServiceStore
	contestRepo contestports.ContestLookupRepository
}

func NewContestAWDServiceQueryService(repo contestports.AWDServiceStore, contestRepo contestports.ContestLookupRepository) *ContestAWDServiceQueryService {
	return &ContestAWDServiceQueryService{repo: repo, contestRepo: contestRepo}
}

func (s *ContestAWDServiceQueryService) ListContestAWDServices(ctx context.Context, contestID int64) ([]ContestAWDServiceResult, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items, err := s.repo.ListContestAWDServicesByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]ContestAWDServiceResult, 0, len(items))
	for i := range items {
		item := items[i]
		snapshot := decodeContestAWDServiceSnapshot(item.ServiceSnapshot)
		resp = append(resp, ContestAWDServiceResult{
			ID:                   item.ID,
			ContestID:            item.ContestID,
			AWDChallengeID:       item.AWDChallengeID,
			Title:                snapshot.Name,
			Category:             snapshot.Category,
			Difficulty:           snapshot.Difficulty,
			DisplayName:          item.DisplayName,
			Order:                item.Order,
			IsVisible:            item.IsVisible,
			ScoreConfig:          contestdomain.ParseAWDCheckerConfig(item.ScoreConfig),
			RuntimeConfig:        sanitizeContestAWDServiceRuntimeConfig(contestdomain.ParseAWDCheckerConfig(item.RuntimeConfig)),
			ValidationState:      string(contestdomain.NormalizeAWDCheckerValidationState(string(item.ValidationState))),
			LastPreviewAt:        item.LastPreviewAt,
			LastPreviewResultRaw: item.LastPreviewResult,
			CreatedAt:            item.CreatedAt,
			UpdatedAt:            item.UpdatedAt,
		})
	}
	return resp, nil
}

type contestAWDServiceSnapshotResult struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
}

func decodeContestAWDServiceSnapshot(raw string) contestAWDServiceSnapshotResult {
	if strings.TrimSpace(raw) == "" {
		return contestAWDServiceSnapshotResult{}
	}
	var snapshot contestAWDServiceSnapshotResult
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return contestAWDServiceSnapshotResult{}
	}
	return snapshot
}

func sanitizeContestAWDServiceRuntimeConfig(runtimeConfig map[string]any) map[string]any {
	if len(runtimeConfig) == 0 {
		return runtimeConfig
	}
	sanitized := make(map[string]any, len(runtimeConfig))
	for key, value := range runtimeConfig {
		if key == "challenge_id" {
			continue
		}
		sanitized[key] = value
	}
	return sanitized
}
