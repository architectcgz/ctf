package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
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

func (s *TagService) CreateTag(req *dto.CreateTagReq) (*dto.TagResp, error) {
	tag := &model.Tag{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}
	if err := s.repo.Create(tag); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.TagRespFromModel(tag), nil
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
