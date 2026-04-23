package queries

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type ImageService struct {
	repo   challengeports.ImageRepository
	config *config.Config
}

func NewImageService(repo challengeports.ImageRepository, config *config.Config) *ImageService {
	return &ImageService{
		repo:   repo,
		config: config,
	}
}

func (s *ImageService) GetImage(id int64) (*dto.ImageResp, error) {
	return s.GetImageWithContext(context.Background(), id)
}

func (s *ImageService) GetImageWithContext(ctx context.Context, id int64) (*dto.ImageResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	image, err := s.repo.FindByIDWithContext(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrImageNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.ImageRespFromModel(image), nil
}

func (s *ImageService) ListImages(query *dto.ImageQuery) (*dto.PageResult, error) {
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
	images, total, err := s.repo.List(query.Name, query.Status, offset, size)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]interface{}, len(images))
	for index, image := range images {
		items[index] = domain.ImageRespFromModel(image)
	}

	return &dto.PageResult{
		List:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}
