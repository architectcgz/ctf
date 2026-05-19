package queries

import (
	"context"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type TagService struct {
	repo challengeports.TagQueryRepository
}

func NewTagService(repo challengeports.TagQueryRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) ListTags(ctx context.Context, tagType string) ([]*challengecontracts.TagResp, error) {
	tags, err := s.repo.List(ctx, tagType)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	result := make([]*challengecontracts.TagResp, len(tags))
	for index, tag := range tags {
		result[index] = domain.ResponseMapper().ToTagRespPtr(tag)
	}
	return result, nil
}

func (s *TagService) GetChallengeTagIDs(ctx context.Context, challengeID int64) ([]int64, error) {
	tags, err := s.repo.FindByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	tagIDs := make([]int64, len(tags))
	for index, tag := range tags {
		tagIDs[index] = tag.ID
	}
	return tagIDs, nil
}
