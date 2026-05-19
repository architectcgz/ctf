package commands

import (
	"context"

	"ctf-platform/internal/apperror"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/challenge/domain"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type TagService struct {
	repo tagCommandRepository
}

type tagCommandRepository interface {
	challengeports.TagCommandRepository
}

func NewTagService(repo tagCommandRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) CreateTag(ctx context.Context, req CreateTagInput) (*challengecontracts.TagResp, error) {
	tag := &challengeentity.Tag{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}
	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return domain.ResponseMapper().ToTagRespPtr(tag), nil
}

func (s *TagService) DeleteTag(ctx context.Context, id int64) error {
	count, err := s.repo.CountChallengesByTagID(ctx, id)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if count > 0 {
		return apperror.ErrConflict.WithCause(nil)
	}
	return s.repo.Delete(ctx, id)
}

func (s *TagService) AttachTags(ctx context.Context, challengeID int64, tagIDs []int64) error {
	tags, err := s.repo.FindByIDs(ctx, tagIDs)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return apperror.ErrNotFound
	}
	return s.repo.AttachTagsInTx(ctx, challengeID, tagIDs)
}

func (s *TagService) DetachTags(ctx context.Context, challengeID int64, tagIDs []int64) error {
	tags, err := s.repo.FindByIDs(ctx, tagIDs)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return apperror.ErrNotFound
	}
	for _, tagID := range tagIDs {
		if err := s.repo.DetachFromChallenge(ctx, challengeID, tagID); err != nil {
			return apperror.ErrInternal.WithCause(err)
		}
	}
	return nil
}
