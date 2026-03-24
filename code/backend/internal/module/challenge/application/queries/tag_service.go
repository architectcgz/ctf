package queries

import (
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

func (s *TagService) ListTags(tagType string) ([]*dto.TagResp, error) {
	tags, err := s.repo.List(tagType)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TagResp, len(tags))
	for index, tag := range tags {
		result[index] = domain.TagRespFromModel(tag)
	}
	return result, nil
}

func (s *TagService) GetChallengeTagIDs(challengeID int64) ([]int64, error) {
	tags, err := s.repo.FindByChallengeID(challengeID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	tagIDs := make([]int64, len(tags))
	for index, tag := range tags {
		tagIDs[index] = tag.ID
	}
	return tagIDs, nil
}
