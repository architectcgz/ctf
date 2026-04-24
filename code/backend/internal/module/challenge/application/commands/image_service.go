package commands

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type ImageService struct {
	repo          challengeports.ImageRepository
	challengeRepo challengeports.ChallengeImageUsageRepository
	runtime       challengeports.ImageRuntime
	logger        *zap.Logger
	baseCtx       context.Context
	cancel        context.CancelFunc
	tasks         sync.WaitGroup
}

func NewImageService(
	repo challengeports.ImageRepository,
	challengeRepo challengeports.ChallengeImageUsageRepository,
	runtime challengeports.ImageRuntime,
	logger *zap.Logger,
) *ImageService {
	if logger == nil {
		logger = zap.NewNop()
	}
	baseCtx, cancel := context.WithCancel(context.Background())
	return &ImageService{
		repo:          repo,
		challengeRepo: challengeRepo,
		runtime:       runtime,
		logger:        logger,
		baseCtx:       baseCtx,
		cancel:        cancel,
	}
}

func (s *ImageService) CreateImage(ctx context.Context, req *dto.CreateImageReq) (*dto.ImageResp, error) {
	existing, err := s.repo.FindByNameTag(ctx, req.Name, req.Tag)
	if err == nil && existing != nil {
		return nil, errcode.ErrImageAlreadyExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var size int64
	if s.runtime != nil {
		imageRef := fmt.Sprintf("%s:%s", req.Name, req.Tag)
		size, err = s.verifyDockerImage(ctx, imageRef)
		if err != nil {
			return nil, errcode.ErrImageNotAccessible.WithCause(err)
		}
	}

	image := &model.Image{
		Name:        req.Name,
		Tag:         req.Tag,
		Description: req.Description,
		Size:        size,
		Status:      model.ImageStatusAvailable,
	}
	if err := s.repo.Create(ctx, image); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("创建镜像", zap.Int64("id", image.ID), zap.String("name", image.Name), zap.String("tag", image.Tag))
	return &dto.ImageResp{
		ID:          image.ID,
		Name:        image.Name,
		Tag:         image.Tag,
		Description: image.Description,
		Size:        image.Size,
		Status:      image.Status,
		CreatedAt:   image.CreatedAt,
		UpdatedAt:   image.UpdatedAt,
	}, nil
}

func (s *ImageService) UpdateImage(ctx context.Context, id int64, req *dto.UpdateImageReq) error {
	image, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrImageNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if req.Description != "" {
		image.Description = req.Description
	}
	if req.Status != "" {
		image.Status = req.Status
	}
	if err := s.repo.Update(ctx, image); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	s.logger.Info("更新镜像", zap.Int64("id", id))
	return nil
}

func (s *ImageService) DeleteImage(ctx context.Context, id int64) error {
	image, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrImageNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	count, err := s.challengeRepo.CountByImageID(ctx, id)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if count > 0 {
		return errcode.ErrImageInUse
	}
	if err := s.repo.DeleteWithContext(ctx, id); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	if s.runtime != nil {
		s.removeImageAsync(fmt.Sprintf("%s:%s", image.Name, image.Tag))
	}
	s.logger.Info("删除镜像", zap.Int64("id", id), zap.String("name", image.Name))
	return nil
}

func (s *ImageService) Close(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	go func() {
		s.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *ImageService) verifyDockerImage(ctx context.Context, imageRef string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.runtime.InspectImageSize(ctx, imageRef)
}

func (s *ImageService) removeImageAsync(imageRef string) {
	s.tasks.Add(1)
	go func() {
		defer s.tasks.Done()
		if err := s.removeImage(imageRef); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("删除 Docker 镜像失败", zap.String("image", imageRef), zap.Error(err))
		}
	}()
}

func (s *ImageService) removeImage(imageRef string) error {
	if s.baseCtx == nil {
		return context.Canceled
	}
	ctx, cancel := context.WithTimeout(s.baseCtx, 30*time.Second)
	defer cancel()
	return s.runtime.RemoveImage(ctx, imageRef)
}
