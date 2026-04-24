package queries

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type TagService struct {
	repo challengeports.TagRepository
}

func NewTagService(repo challengeports.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) ListTagsWithContext(ctx context.Context, tagType string) ([]*dto.TagResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	tags, err := s.repo.ListWithContext(ctx, tagType)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TagResp, len(tags))
	for index, tag := range tags {
		result[index] = domain.TagRespFromModel(tag)
	}
	return result, nil
}

func (s *TagService) GetChallengeTagIDsWithContext(ctx context.Context, challengeID int64) ([]int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	tags, err := s.repo.FindByChallengeIDWithContext(ctx, challengeID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	tagIDs := make([]int64, len(tags))
	for index, tag := range tags {
		tagIDs[index] = tag.ID
	}
	return tagIDs, nil
}
