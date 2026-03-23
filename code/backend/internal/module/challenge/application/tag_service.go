package application

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type TagService struct {
	repo TagRepository
}

func NewTagService(repo TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) CreateTag(req *dto.CreateTagReq) (*dto.TagResp, error) {
	tag := &model.Tag{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}

	if err := s.repo.Create(tag); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return toTagResp(tag), nil
}

func (s *TagService) ListTags(tagType string) ([]*dto.TagResp, error) {
	tags, err := s.repo.List(tagType)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TagResp, len(tags))
	for i, tag := range tags {
		result[i] = toTagResp(tag)
	}
	return result, nil
}

func (s *TagService) DeleteTag(id int64) error {
	count, err := s.repo.CountChallengesByTagID(id)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if count > 0 {
		return errcode.ErrConflict.WithCause(nil)
	}

	return s.repo.Delete(id)
}

func (s *TagService) AttachTags(challengeID int64, tagIDs []int64) error {
	tags, err := s.repo.FindByIDs(tagIDs)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return errcode.ErrNotFound
	}

	return s.repo.AttachTagsInTx(challengeID, tagIDs)
}

func (s *TagService) DetachTags(challengeID int64, tagIDs []int64) error {
	tags, err := s.repo.FindByIDs(tagIDs)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return errcode.ErrNotFound
	}

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
		ID:          tag.ID,
		Name:        tag.Name,
		Type:        tag.Type,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
	}
}
