package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/dto"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type AWDChallengeQueryService struct {
	repo challengeports.AWDChallengeQueryRepository
}

func NewAWDChallengeQueryService(repo challengeports.AWDChallengeQueryRepository) *AWDChallengeQueryService {
	return &AWDChallengeQueryService{repo: repo}
}

func (s *AWDChallengeQueryService) GetChallenge(ctx context.Context, id int64) (*challengecontracts.AWDChallengeResp, error) {
	item, err := s.repo.FindAWDChallengeByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrAWDChallengeNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDChallengeRespFromModel(item), nil
}

func (s *AWDChallengeQueryService) ListChallenges(ctx context.Context, req ListAWDChallengesInput) (*challengecontracts.AWDChallengePageResp, error) {
	query := &dto.AWDChallengeQuery{
		Keyword:     req.Keyword,
		ServiceType: req.ServiceType,
		Status:      req.Status,
		Page:        req.Page,
		Size:        req.Size,
	}

	items, total, err := s.repo.ListAWDChallenges(ctx, query)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	page := 1
	size := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.Size > 0 {
		size = req.Size
	}
	resp := &challengecontracts.AWDChallengePageResp{
		Items: make([]*challengecontracts.AWDChallengeResp, 0, len(items)),
		Total: total,
		Page:  page,
		Size:  size,
	}
	for _, item := range items {
		resp.Items = append(resp.Items, domain.AWDChallengeRespFromModel(item))
	}
	return resp, nil
}
