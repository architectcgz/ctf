package queries

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) GetContestChallenges(ctx context.Context, userID, contestID int64) ([]*dto.ContestChallengeInfo, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestChallengeVisible
	}
	if contest.Mode == model.ContestModeAWD {
		return s.getAWDContestChallenges(ctx, userID, contestID)
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, true)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if len(challenges) == 0 {
		return []*dto.ContestChallengeInfo{}, nil
	}
	servicesByChallenge := make(map[int64]model.ContestAWDService)
	if s.awdRepo != nil {
		services, listErr := s.awdRepo.ListContestAWDServicesByContest(ctx, contestID)
		if listErr != nil {
			return nil, errcode.ErrInternal.WithCause(listErr)
		}
		for i := range services {
			item := services[i]
			servicesByChallenge[item.ChallengeID] = item
		}
	}

	challengeIDs := make([]int64, 0, len(challenges))
	for _, item := range challenges {
		challengeIDs = append(challengeIDs, item.ChallengeID)
	}

	solvedMap, err := s.challengeRepo.BatchGetSolvedStatus(ctx, userID, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	solvedCountMap, err := s.challengeRepo.BatchGetSolvedCount(ctx, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeInfo, 0, len(challenges))
	for _, item := range challenges {
		challenge, findErr := s.challengeRepo.FindByIDWithContext(ctx, item.ChallengeID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		resp := &dto.ContestChallengeInfo{
			ID:          item.ID,
			ChallengeID: item.ChallengeID,
			Title:       challenge.Title,
			Category:    challenge.Category,
			Difficulty:  challenge.Difficulty,
			Points:      item.Points,
			Order:       item.Order,
			SolvedCount: solvedCountMap[item.ChallengeID],
			IsSolved:    solvedMap[item.ChallengeID],
		}
		if service, ok := servicesByChallenge[item.ChallengeID]; ok {
			resp.AWDServiceID = &service.ID
		}
		result = append(result, resp)
	}
	return result, nil
}

func (s *ChallengeService) getAWDContestChallenges(ctx context.Context, userID, contestID int64) ([]*dto.ContestChallengeInfo, error) {
	if s.awdRepo == nil {
		return []*dto.ContestChallengeInfo{}, nil
	}

	services, err := s.awdRepo.ListContestAWDServicesByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	visibleServices := make([]model.ContestAWDService, 0, len(services))
	challengeIDs := make([]int64, 0, len(services))
	for _, service := range services {
		if !service.IsVisible {
			continue
		}
		visibleServices = append(visibleServices, service)
		challengeIDs = append(challengeIDs, service.ChallengeID)
	}
	if len(visibleServices) == 0 {
		return []*dto.ContestChallengeInfo{}, nil
	}

	solvedMap, err := s.challengeRepo.BatchGetSolvedStatus(ctx, userID, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	solvedCountMap, err := s.challengeRepo.BatchGetSolvedCount(ctx, challengeIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.ContestChallengeInfo, 0, len(visibleServices))
	for _, service := range visibleServices {
		snapshot, decodeErr := model.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
		if decodeErr != nil {
			return nil, errcode.ErrInternal.WithCause(decodeErr)
		}
		result = append(result, &dto.ContestChallengeInfo{
			ID:           service.ID,
			ChallengeID:  service.ChallengeID,
			AWDServiceID: &service.ID,
			Title:        firstAWDServiceValue(service.DisplayName, snapshot.Name),
			Category:     snapshot.Category,
			Difficulty:   snapshot.Difficulty,
			Points:       parseAWDServicePoints(service.ScoreConfig),
			Order:        service.Order,
			SolvedCount:  solvedCountMap[service.ChallengeID],
			IsSolved:     solvedMap[service.ChallengeID],
		})
	}
	return result, nil
}

func parseAWDServicePoints(scoreConfig string) int {
	if strings.TrimSpace(scoreConfig) == "" {
		return 0
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(scoreConfig), &payload); err != nil {
		return 0
	}
	switch typed := payload["points"].(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	default:
		return 0
	}
}

func firstAWDServiceValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
