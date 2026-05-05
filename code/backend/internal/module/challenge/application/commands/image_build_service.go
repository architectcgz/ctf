package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ImageBuildConfig struct {
	Registry         string
	PollInterval     time.Duration
	BatchSize        int
	BuildTimeout     time.Duration
	BuildConcurrency int
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
	repo     imageBuildRepository
	builder  challengeports.DockerImageBuilder
	verifier challengeports.RegistryVerifier
	config   ImageBuildConfig
	logger   *zap.Logger
	cancel   context.CancelFunc
	tasks    sync.WaitGroup
}

type ImageBuildOption func(*ImageBuildService)

func WithImageBuildDockerBuilder(builder challengeports.DockerImageBuilder) ImageBuildOption {
	return func(s *ImageBuildService) {
		s.builder = builder
	}
}

func WithImageBuildRegistryVerifier(verifier challengeports.RegistryVerifier) ImageBuildOption {
	return func(s *ImageBuildService) {
		s.verifier = verifier
	}
}

func WithImageBuildLogger(logger *zap.Logger) ImageBuildOption {
	return func(s *ImageBuildService) {
		if logger != nil {
			s.logger = logger
		}
	}
}

func NewImageBuildService(repo imageBuildRepository, config ImageBuildConfig, options ...ImageBuildOption) *ImageBuildService {
	if config.PollInterval <= 0 {
		config.PollInterval = 2 * time.Second
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 1
	}
	if config.BuildTimeout <= 0 {
		config.BuildTimeout = 10 * time.Minute
	}
	if config.BuildConcurrency <= 0 {
		config.BuildConcurrency = 1
	}
	service := &ImageBuildService{
		repo:   repo,
		config: config,
		logger: zap.NewNop(),
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
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

func (s *ImageBuildService) StartBackgroundTasks(ctx context.Context) {
	if s == nil || ctx == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	loopCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.tasks.Add(1)
	go func() {
		defer s.tasks.Done()
		s.RunBuildLoop(loopCtx)
	}()
}

func (s *ImageBuildService) Close(ctx context.Context) error {
	if s == nil {
		return nil
	}
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

func (s *ImageBuildService) RunBuildLoop(ctx context.Context) {
	if s == nil {
		return
	}
	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()
	for {
		if err := s.ProcessPendingBuildJobs(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("process image build jobs failed", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *ImageBuildService) ProcessPendingBuildJobs(ctx context.Context) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("image build service is not configured")
	}
	jobs, err := s.repo.ListPendingImageBuildJobs(ctx, s.config.BatchSize)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	sem := make(chan struct{}, s.config.BuildConcurrency)
	var wg sync.WaitGroup
	errCh := make(chan error, len(jobs))
	for _, job := range jobs {
		jobID := job.ID
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := s.ProcessImageBuildJob(ctx, jobID); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ImageBuildService) ProcessImageBuildJob(ctx context.Context, jobID int64) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("image build service is not configured")
	}
	if s.builder == nil {
		return fmt.Errorf("docker image builder is not configured")
	}
	if s.verifier == nil {
		return fmt.Errorf("registry verifier is not configured")
	}

	startedAt := time.Now()
	started, err := s.repo.TryStartImageBuildJob(ctx, jobID, startedAt)
	if err != nil {
		return err
	}
	if !started {
		return nil
	}

	job, err := s.repo.FindImageBuildJobByID(ctx, jobID)
	if err != nil {
		return err
	}
	image, err := s.findImageByJobTargetRef(ctx, job.TargetRef)
	if err != nil {
		return err
	}
	if err := s.updateImageBuildStatus(ctx, image, model.ImageStatusBuilding, "", ""); err != nil {
		return err
	}

	buildCtx, cancel := context.WithTimeout(ctx, s.config.BuildTimeout)
	defer cancel()

	if err := s.builder.Build(buildCtx, job.ContextPath, job.DockerfilePath, job.TargetRef); err != nil {
		return s.failImageBuildJob(ctx, job, image, err)
	}
	if err := s.builder.Push(buildCtx, job.TargetRef); err != nil {
		return s.failImageBuildJob(ctx, job, image, err)
	}
	now := time.Now()
	job.Status = model.ImageBuildJobStatusPushed
	job.UpdatedAt = now
	if err := s.repo.UpdateImageBuildJob(ctx, job); err != nil {
		return err
	}
	if err := s.updateImageBuildStatus(ctx, image, model.ImageStatusPushed, "", ""); err != nil {
		return err
	}

	job.Status = model.ImageBuildJobStatusVerifying
	job.UpdatedAt = time.Now()
	if err := s.repo.UpdateImageBuildJob(ctx, job); err != nil {
		return err
	}
	if err := s.updateImageBuildStatus(ctx, image, model.ImageStatusVerifying, "", ""); err != nil {
		return err
	}

	digest, err := s.verifier.CheckManifest(buildCtx, job.TargetRef)
	if err != nil {
		return s.failImageBuildJob(ctx, job, image, err)
	}
	if err := s.builder.Pull(buildCtx, job.TargetRef); err != nil {
		return s.failImageBuildJob(ctx, job, image, err)
	}
	inspect, err := s.builder.Inspect(buildCtx, job.TargetRef)
	if err != nil {
		return s.failImageBuildJob(ctx, job, image, err)
	}

	finishedAt := time.Now()
	job.Status = model.ImageBuildJobStatusAvailable
	job.TargetDigest = digest
	job.FinishedAt = &finishedAt
	job.ErrorSummary = ""
	job.UpdatedAt = finishedAt
	if err := s.repo.UpdateImageBuildJob(ctx, job); err != nil {
		return err
	}
	image.Size = inspect.Size
	return s.updateImageBuildStatus(ctx, image, model.ImageStatusAvailable, digest, "")
}

func (s *ImageBuildService) failImageBuildJob(ctx context.Context, job *model.ImageBuildJob, image *model.Image, cause error) error {
	summary := strings.TrimSpace(cause.Error())
	finishedAt := time.Now()
	job.Status = model.ImageBuildJobStatusFailed
	job.ErrorSummary = summary
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt
	if err := s.repo.UpdateImageBuildJob(ctx, job); err != nil {
		return err
	}
	if err := s.updateImageBuildStatus(ctx, image, model.ImageStatusFailed, "", summary); err != nil {
		return err
	}
	return cause
}

func (s *ImageBuildService) findImageByJobTargetRef(ctx context.Context, targetRef string) (*model.Image, error) {
	name, tag, err := domain.SplitImageRef(targetRef)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByNameTag(ctx, name, tag)
}

func (s *ImageBuildService) updateImageBuildStatus(ctx context.Context, image *model.Image, status, digest, lastError string) error {
	image.Status = status
	image.SourceType = model.ImageSourceTypePlatformBuild
	image.Digest = digest
	image.LastError = lastError
	if status == model.ImageStatusAvailable {
		now := time.Now()
		image.VerifiedAt = &now
	} else {
		image.VerifiedAt = nil
	}
	return s.repo.Update(ctx, image)
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
