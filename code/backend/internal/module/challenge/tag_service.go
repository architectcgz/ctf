package challenge

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"

	"gorm.io/gorm"
)

type TagService struct {
	repo *TagRepository
}

func NewTagService(repo *TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) CreateTag(req *dto.CreateTagReq) (*dto.TagResp, error) {
	tag := &model.Tag{
		Name:      req.Name,
		Dimension: req.Dimension,
	}

	if err := s.repo.Create(tag); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return toTagResp(tag), nil
}

func (s *TagService) ListTags(dimension string) ([]*dto.TagResp, error) {
	tags, err := s.repo.List(dimension)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TagResp, len(tags))
	for i, tag := range tags {
		result[i] = toTagResp(tag)
	}
	return result, nil
}

func (s *TagService) AttachTags(challengeID int64, tagIDs []int64) error {
	for _, tagID := range tagIDs {
		if _, err := s.repo.FindByID(tagID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return errcode.ErrNotFound("标签")
			}
			return errcode.ErrInternal.WithCause(err)
		}

		if err := s.repo.AttachToChallenge(challengeID, tagID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	return nil
}

func (s *TagService) DetachTags(challengeID int64, tagIDs []int64) error {
	for _, tagID := range tagIDs {
		if err := s.repo.DetachFromChallenge(challengeID, tagID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	return nil
}

func (s *TagService) GetChallengeTagIDs(challengeID int64) ([]int64, error) {
	tags, err := s.repo.FindByChallengeID(challengeID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	tagIDs := make([]int64, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}
	return tagIDs, nil
}

func toTagResp(tag *model.Tag) *dto.TagResp {
	return &dto.TagResp{
		ID:        tag.ID,
		Name:      tag.Name,
		Dimension: tag.Dimension,
		CreatedAt: tag.CreatedAt,
	}
}
