package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ImageBuildConfig struct {
	Registry string
}

type CreatePlatformBuildJobRequest struct {
	ChallengeMode  string
	PackageSlug    string
	SuggestedTag   string
	SourceDir      string
	DockerfilePath string
	ContextPath    string
	CreatedBy      int64
}

type CreatePlatformBuildJobResult struct {
	ImageID   int64
	JobID     int64
	TargetRef string
}

type imageBuildRepository interface {
	challengeports.ImageCommandRepository
	challengeports.ImageBuildJobRepository
}

type ImageBuildService struct {
	repo   imageBuildRepository
	config ImageBuildConfig
}

func NewImageBuildService(repo imageBuildRepository, config ImageBuildConfig) *ImageBuildService {
	return &ImageBuildService{repo: repo, config: config}
}

func (s *ImageBuildService) CreatePlatformBuildJob(
	ctx context.Context,
	req CreatePlatformBuildJobRequest,
) (*CreatePlatformBuildJobResult, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}

	tag := strings.TrimSpace(req.SuggestedTag)
	if tag == "" {
		tag = "latest"
	}
	targetRef, err := domain.BuildPlatformImageRef(s.config.Registry, req.ChallengeMode, req.PackageSlug, tag)
	if err != nil {
		return nil, err
	}
	name, imageTag, err := domain.SplitImageRef(targetRef)
	if err != nil {
		return nil, err
	}

	image, err := s.findOrCreatePendingPlatformBuildImage(ctx, name, imageTag, req.PackageSlug)
	if err != nil {
		return nil, err
	}

	createdBy := req.CreatedBy
	job := &model.ImageBuildJob{
		SourceType:     model.ImageSourceTypePlatformBuild,
		ChallengeMode:  strings.TrimSpace(req.ChallengeMode),
		PackageSlug:    strings.TrimSpace(req.PackageSlug),
		SourceDir:      strings.TrimSpace(req.SourceDir),
		DockerfilePath: strings.TrimSpace(req.DockerfilePath),
		ContextPath:    strings.TrimSpace(req.ContextPath),
		TargetRef:      targetRef,
		Status:         model.ImageBuildJobStatusPending,
	}
	if createdBy > 0 {
		job.CreatedBy = &createdBy
	}
	if err := s.repo.CreateImageBuildJob(ctx, job); err != nil {
		return nil, err
	}

	image.Status = model.ImageStatusPending
	image.SourceType = model.ImageSourceTypePlatformBuild
	image.BuildJobID = &job.ID
	image.LastError = ""
	image.Digest = ""
	image.VerifiedAt = nil
	if err := s.repo.Update(ctx, image); err != nil {
		return nil, err
	}

	return &CreatePlatformBuildJobResult{
		ImageID:   image.ID,
		JobID:     job.ID,
		TargetRef: targetRef,
	}, nil
}

func (s *ImageBuildService) findOrCreatePendingPlatformBuildImage(
	ctx context.Context,
	name string,
	tag string,
	packageSlug string,
) (*model.Image, error) {
	image, err := s.repo.FindByNameTag(ctx, name, tag)
	switch {
	case err == nil:
		return image, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		image = &model.Image{
			Name:        name,
			Tag:         tag,
			Description: fmt.Sprintf("Built from challenge pack %s", packageSlug),
			Status:      model.ImageStatusPending,
			SourceType:  model.ImageSourceTypePlatformBuild,
		}
		if err := s.repo.Create(ctx, image); err != nil {
			return nil, err
		}
		return image, nil
	default:
		return nil, err
	}
}
