package commands

import (
	"context"

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
	return s.CreateTagWithContext(context.Background(), req)
}

func (s *TagService) CreateTagWithContext(ctx context.Context, req *dto.CreateTagReq) (*dto.TagResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	tag := &model.Tag{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	}
	if err := s.repo.CreateWithContext(ctx, tag); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.TagRespFromModel(tag), nil
}

func (s *TagService) DeleteTag(id int64) error {
	return s.DeleteTagWithContext(context.Background(), id)
}

func (s *TagService) DeleteTagWithContext(ctx context.Context, id int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	count, err := s.repo.CountChallengesByTagIDWithContext(ctx, id)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if count > 0 {
		return errcode.ErrConflict.WithCause(nil)
	}
	return s.repo.DeleteWithContext(ctx, id)
}

func (s *TagService) AttachTags(challengeID int64, tagIDs []int64) error {
	return s.AttachTagsWithContext(context.Background(), challengeID, tagIDs)
}

func (s *TagService) AttachTagsWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tags, err := s.repo.FindByIDsWithContext(ctx, tagIDs)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return errcode.ErrNotFound
	}
	return s.repo.AttachTagsInTxWithContext(ctx, challengeID, tagIDs)
}

func (s *TagService) DetachTags(challengeID int64, tagIDs []int64) error {
	return s.DetachTagsWithContext(context.Background(), challengeID, tagIDs)
}

func (s *TagService) DetachTagsWithContext(ctx context.Context, challengeID int64, tagIDs []int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tags, err := s.repo.FindByIDsWithContext(ctx, tagIDs)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if len(tags) != len(tagIDs) {
		return errcode.ErrNotFound
	}
	for _, tagID := range tagIDs {
		if err := s.repo.DetachFromChallengeWithContext(ctx, challengeID, tagID); err != nil {
			return errcode.ErrInternal.WithCause(err)
		}
	}
	return nil
}
