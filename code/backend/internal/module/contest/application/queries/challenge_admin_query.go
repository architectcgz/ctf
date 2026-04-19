package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) ListAdminChallenges(ctx context.Context, contestID int64) ([]*dto.ContestChallengeResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	challenges, err := s.repo.ListChallenges(ctx, contestID, false)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
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

	result := make([]*dto.ContestChallengeResp, len(challenges))
	for i, item := range challenges {
		challenge, findErr := s.challengeRepo.FindByID(item.ChallengeID)
		if findErr != nil {
			return nil, errcode.ErrInternal.WithCause(findErr)
		}
		resp := contestdomain.ContestChallengeRespFromModel(item, challenge)
		if service, ok := servicesByChallenge[item.ChallengeID]; ok {
			resp.AWDServiceID = &service.ID
			resp.AWDTemplateID = service.TemplateID
			resp.AWDServiceDisplayName = service.DisplayName
		}
		result[i] = resp
	}
	return result, nil
}
