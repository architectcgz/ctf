package queries

import (
	"context"
	"errors"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type ImageService struct {
	repo   challengeports.ImageQueryRepository
	config *config.Config
}

func NewImageService(repo challengeports.ImageQueryRepository, config *config.Config) *ImageService {
	return &ImageService{
		repo:   repo,
		config: config,
	}
}

func (s *ImageService) GetImage(ctx context.Context, id int64) (*dto.ImageResp, error) {
	image, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, challengeports.ErrChallengeImageNotFound) {
			return nil, errcode.ErrImageNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.ImageRespFromModel(image), nil
}

func (s *ImageService) ListImages(ctx context.Context, query ListImagesInput) (*dto.PageResult[*dto.ImageResp], error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	size := query.Size
	if size < 1 {
		size = s.config.Pagination.DefaultPageSize
	}
	if size > s.config.Pagination.MaxPageSize {
		size = s.config.Pagination.MaxPageSize
	}

	offset := (page - 1) * size
	images, total, err := s.repo.List(ctx, query.Name, query.Status, offset, size)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]*dto.ImageResp, len(images))
	for index, image := range images {
		items[index] = domain.ImageRespFromModel(image)
	}

	return &dto.PageResult[*dto.ImageResp]{
		List:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}
