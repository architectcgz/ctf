package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type awdChallengeImportTxRunner struct {
	repo       *challengeinfra.Repository
	imageBuild *challengecmd.ImageBuildService
}

func NewAWDChallengeImportTxRunner(
	repo *challengeinfra.Repository,
	imageBuild *challengecmd.ImageBuildService,
) challengeports.AWDChallengeImportTxRunner {
	if repo == nil {
		return nil
	}
	return &awdChallengeImportTxRunner{
		repo:       repo,
		imageBuild: imageBuild,
	}
}

func (r *awdChallengeImportTxRunner) WithinAWDChallengeImportTransaction(
	ctx context.Context,
	fn func(store challengeports.AWDChallengeImportTxStore) error,
) error {
	if r == nil || r.repo == nil {
		return fmt.Errorf("awd challenge import tx runner is not configured")
	}
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &awdChallengeImportTxStore{
			rawRepo:    txRepo,
			imageBuild: r.imageBuild,
		}
		return fn(store)
	})
}

type awdChallengeImportTxStore struct {
	rawRepo    *challengeinfra.Repository
	imageBuild *challengecmd.ImageBuildService
}

func (s *awdChallengeImportTxStore) tx(ctx context.Context) *gorm.DB {
	if s == nil || s.rawRepo == nil {
		return nil
	}
	return s.rawRepo.DB(ctx)
}

func (s *awdChallengeImportTxStore) RejectImportedAWDChallengeSlugConflict(ctx context.Context, slug string) error {
	normalizedSlug := strings.TrimSpace(slug)
	if normalizedSlug == "" {
		return nil
	}

	var existing model.AWDChallenge
	err := s.tx(ctx).Unscoped().
		Select("id", "slug", "name").
		Where("slug = ?", normalizedSlug).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return fmt.Errorf("check imported awd challenge slug %s: %w", normalizedSlug, err)
	default:
		message := fmt.Sprintf("AWD 题目 slug %s 已被已有题目占用，请改用题目编辑入口更新", normalizedSlug)
		return errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}
}

func (s *awdChallengeImportTxStore) CreateImportedAWDChallenge(
	ctx context.Context,
	challenge *model.AWDChallenge,
) error {
	return s.rawRepo.CreateAWDChallenge(ctx, challenge)
}

func (s *awdChallengeImportTxStore) ResolvePlatformBuildImage(
	ctx context.Context,
	req challengeports.ImportedPlatformBuildImageRequest,
) (*challengeports.ImportedImageResolution, error) {
	if s.imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := s.imageBuild.CreatePlatformBuildJobInTx(
		ctx,
		&runtimeImageBuildTxStore{tx: s.tx(ctx)},
		challengecmd.CreatePlatformBuildJobRequest{
			ChallengeMode:  req.ChallengeMode,
			PackageSlug:    req.PackageSlug,
			SuggestedTag:   req.SuggestedTag,
			SourceDir:      req.SourceDir,
			DockerfilePath: req.DockerfilePath,
			ContextPath:    req.ContextPath,
			CreatedBy:      req.CreatedBy,
		},
	)
	if err != nil {
		return nil, err
	}
	return &challengeports.ImportedImageResolution{
		ImageID:  result.ImageID,
		ImageRef: result.TargetRef,
	}, nil
}

func (s *awdChallengeImportTxStore) ResolveExternalImage(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	if s.imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := s.imageBuild.VerifyExternalImageRefInTx(
		ctx,
		&runtimeImageBuildTxStore{tx: s.tx(ctx)},
		packageSlug,
		imageRef,
	)
	if err != nil {
		return nil, err
	}
	return &challengeports.ImportedImageResolution{
		ImageID:  result.ImageID,
		ImageRef: result.ImageRef,
	}, nil
}

func (s *awdChallengeImportTxStore) ResolveExistingImageRef(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	baseStore := challengeImportTxStore{rawRepo: s.rawRepo}
	return baseStore.ResolveExistingImageRef(ctx, packageSlug, imageRef)
}
