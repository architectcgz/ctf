package challenge

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type ImageService struct {
	repo          *ImageRepository
	challengeRepo *Repository
	dockerClient  *client.Client
	config        *config.Config
	logger        *zap.Logger
}

func NewImageService(
	repo *ImageRepository,
	challengeRepo *Repository,
	dockerClient *client.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *ImageService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ImageService{
		repo:          repo,
		challengeRepo: challengeRepo,
		dockerClient:  dockerClient,
		config:        cfg,
		logger:        logger,
	}
}

func (s *ImageService) CreateImage(req *dto.CreateImageReq) (*dto.ImageResp, error) {
	// 检查镜像是否已存在
	existing, err := s.repo.FindByNameTag(req.Name, req.Tag)
	if err == nil && existing != nil {
		return nil, errcode.ErrImageAlreadyExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	// 验证 Docker 镜像是否存在（如果 Docker 客户端可用）
	var size int64
	if s.dockerClient != nil {
		imageRef := fmt.Sprintf("%s:%s", req.Name, req.Tag)
		size, err = s.verifyDockerImage(imageRef)
		if err != nil {
			return nil, errcode.ErrImageNotAccessible.WithCause(err)
		}
	}

	img := &model.Image{
		Name:        req.Name,
		Tag:         req.Tag,
		Description: req.Description,
		Size:        size,
		Status:      model.ImageStatusAvailable,
	}

	if err := s.repo.Create(img); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("创建镜像", zap.Int64("id", img.ID), zap.String("name", img.Name), zap.String("tag", img.Tag))
	return toImageResp(img), nil
}

func (s *ImageService) GetImage(id int64) (*dto.ImageResp, error) {
	img, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrImageNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return toImageResp(img), nil
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

	offset := (page - 1) * size
	images, total, err := s.repo.List(query.Name, query.Status, offset, size)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	items := make([]interface{}, len(images))
	for i, img := range images {
		items[i] = toImageResp(img)
	}

	return &dto.PageResult{
		List:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

func (s *ImageService) UpdateImage(id int64, req *dto.UpdateImageReq) error {
	img, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrImageNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if req.Description != "" {
		img.Description = req.Description
	}
	if req.Status != "" {
		img.Status = req.Status
	}

	if err := s.repo.Update(img); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("更新镜像", zap.Int64("id", id))
	return nil
}

func (s *ImageService) DeleteImage(id int64) error {
	img, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrImageNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	count, err := s.challengeRepo.CountByImageID(id)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if count > 0 {
		return errcode.ErrImageInUse
	}

	// 删除数据库记录
	if err := s.repo.Delete(id); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	// 尝试删除 Docker 镜像（非阻塞，仅当客户端可用时）
	if s.dockerClient != nil {
		imageRef := fmt.Sprintf("%s:%s", img.Name, img.Tag)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if _, err := s.dockerClient.ImageRemove(ctx, imageRef, image.RemoveOptions{}); err != nil {
				s.logger.Warn("删除 Docker 镜像失败", zap.String("image", imageRef), zap.Error(err))
			}
		}()
	}

	s.logger.Info("删除镜像", zap.Int64("id", id), zap.String("name", img.Name))
	return nil
}

func (s *ImageService) verifyDockerImage(imageRef string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inspect, _, err := s.dockerClient.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return 0, err
	}
	return inspect.Size, nil
}

func toImageResp(img *model.Image) *dto.ImageResp {
	return &dto.ImageResp{
		ID:          img.ID,
		Name:        img.Name,
		Tag:         img.Tag,
		Description: img.Description,
		Size:        img.Size,
		Status:      img.Status,
		CreatedAt:   img.CreatedAt,
		UpdatedAt:   img.UpdatedAt,
	}
}
