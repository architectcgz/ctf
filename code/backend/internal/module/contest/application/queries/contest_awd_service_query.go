package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type ContestAWDServiceQueryService struct {
	repo        contestports.AWDRepository
	contestRepo contestports.ContestLookupRepository
}

func NewContestAWDServiceQueryService(repo contestports.AWDRepository, contestRepo contestports.ContestLookupRepository) *ContestAWDServiceQueryService {
	return &ContestAWDServiceQueryService{repo: repo, contestRepo: contestRepo}
}

func (s *ContestAWDServiceQueryService) ListContestAWDServices(ctx context.Context, contestID int64) ([]*dto.ContestAWDServiceResp, error) {
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

	resp := make([]*dto.ContestAWDServiceResp, 0, len(items))
	for i := range items {
		item := items[i]
		resp = append(resp, contestdomain.ContestAWDServiceRespFromModel(&item))
	}
	return resp, nil
}
