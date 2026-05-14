package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	"ctf-platform/internal/module/challenge/domain"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type runtimeImageBuildTxStore struct {
	tx *gorm.DB
}

func (s *runtimeImageBuildTxStore) FindByNameTag(ctx context.Context, name string, tag string) (*model.Image, error) {
	if s == nil || s.tx == nil {
		return nil, fmt.Errorf("image build transaction is not configured")
	}
	var image model.Image
	err := s.tx.WithContext(ctx).Unscoped().
		Where("name = ? AND tag = ?", name, tag).
		First(&image).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeImageNotFound
	}
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (s *runtimeImageBuildTxStore) CreateImage(ctx context.Context, image *model.Image) error {
	if s == nil || s.tx == nil {
		return fmt.Errorf("image build transaction is not configured")
	}
	return s.tx.WithContext(ctx).Create(image).Error
}

func (s *runtimeImageBuildTxStore) CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	if s == nil || s.tx == nil {
		return fmt.Errorf("image build transaction is not configured")
	}
	return s.tx.WithContext(ctx).Create(job).Error
}

func (s *runtimeImageBuildTxStore) UpdateImage(ctx context.Context, image *model.Image, updates map[string]any) error {
	if s == nil || s.tx == nil {
		return fmt.Errorf("image build transaction is not configured")
	}
	return s.tx.WithContext(ctx).Unscoped().Model(image).Updates(updates).Error
}

type challengeImportTxRunner struct {
	repo       *challengeinfra.Repository
	imageBuild *challengecmd.ImageBuildService
}

func NewChallengeImportTxRunner(
	repo *challengeinfra.Repository,
	imageBuild *challengecmd.ImageBuildService,
) challengeports.ChallengeImportTxRunner {
	if repo == nil {
		return nil
	}
	return &challengeImportTxRunner{
		repo:       repo,
		imageBuild: imageBuild,
	}
}

func (r *challengeImportTxRunner) WithinChallengeImportTransaction(
	ctx context.Context,
	fn func(store challengeports.ChallengeImportTxStore) error,
) error {
	if r == nil || r.repo == nil {
		return fmt.Errorf("challenge import tx runner is not configured")
	}
	return r.repo.WithinTransaction(ctx, func(txRepo *challengeinfra.Repository) error {
		store := &challengeImportTxStore{
			rawRepo:    txRepo,
			imageBuild: r.imageBuild,
		}
		return fn(store)
	})
}

type challengeImportTxStore struct {
	rawRepo    *challengeinfra.Repository
	imageBuild *challengecmd.ImageBuildService
}

func (s *challengeImportTxStore) tx() *gorm.DB {
	if s == nil || s.rawRepo == nil {
		return nil
	}
	return s.rawRepo.DB()
}

func (s *challengeImportTxStore) RejectImportedChallengeSlugConflict(ctx context.Context, packageSlug string) error {
	slug := strings.TrimSpace(packageSlug)
	if slug == "" {
		return nil
	}

	var existing model.Challenge
	err := s.tx().WithContext(ctx).Unscoped().
		Select("id", "title", "package_slug").
		Where("package_slug = ?", slug).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return fmt.Errorf("check imported challenge slug %s: %w", slug, err)
	default:
		message := fmt.Sprintf("题目 slug %s 已被已有题目占用，请改用题目编辑入口更新", slug)
		return errcode.New(errcode.ErrConflict.Code, message, errcode.ErrConflict.HTTPStatus)
	}
}

func (s *challengeImportTxStore) FindLegacyChallengeForImportedPackageCreate(
	ctx context.Context,
	title string,
	category string,
) (*model.Challenge, bool, error) {
	var challenge model.Challenge
	err := s.tx().WithContext(ctx).Unscoped().
		Where("(package_slug IS NULL OR package_slug = '') AND title = ? AND category = ?", title, category).
		First(&challenge).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, false, nil
	case err != nil:
		return nil, false, fmt.Errorf("find imported challenge %s: %w", strings.TrimSpace(title), err)
	default:
		return &challenge, true, nil
	}
}

func (s *challengeImportTxStore) CreateImportedChallenge(ctx context.Context, challenge *model.Challenge) error {
	if challenge == nil {
		return fmt.Errorf("challenge is nil")
	}
	return s.tx().WithContext(ctx).Create(challenge).Error
}

func (s *challengeImportTxStore) UpdateImportedChallenge(
	ctx context.Context,
	challenge *model.Challenge,
	updates map[string]any,
) error {
	if challenge == nil {
		return fmt.Errorf("challenge is nil")
	}
	return s.tx().WithContext(ctx).Unscoped().Model(challenge).Updates(updates).Error
}

func (s *challengeImportTxStore) ClearPublishCheckJobs(ctx context.Context, challengeID int64) error {
	return s.tx().WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Delete(&model.ChallengePublishCheckJob{}).Error
}

func (s *challengeImportTxStore) ReplaceImportedHints(
	ctx context.Context,
	challengeID int64,
	hints []model.ChallengeHint,
) error {
	if err := s.tx().WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Delete(&model.ChallengeHint{}).Error; err != nil {
		return fmt.Errorf("delete hints for challenge %d: %w", challengeID, err)
	}
	if len(hints) == 0 {
		return nil
	}
	return s.tx().WithContext(ctx).Create(&hints).Error
}

func (s *challengeImportTxStore) ApplyImportedFlagUpdates(
	ctx context.Context,
	challengeID int64,
	updates map[string]any,
) error {
	return s.tx().WithContext(ctx).
		Model(&model.Challenge{}).
		Where("id = ?", challengeID).
		Updates(updates).Error
}

func (s *challengeImportTxStore) NextChallengePackageRevisionNo(ctx context.Context, challengeID int64) (int, error) {
	var latest model.ChallengePackageRevision
	err := s.tx().WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Order("revision_no DESC, id DESC").
		First(&latest).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 1, nil
	case err != nil:
		return 0, err
	default:
		return latest.RevisionNo + 1, nil
	}
}

func (s *challengeImportTxStore) CreateImportedPackageRevision(
	ctx context.Context,
	revision *model.ChallengePackageRevision,
) error {
	return s.rawRepo.CreateChallengePackageRevision(ctx, revision)
}

func (s *challengeImportTxStore) UpsertImportedTopology(
	ctx context.Context,
	topology *model.ChallengeTopology,
) error {
	return s.rawRepo.UpsertChallengeTopology(ctx, topology)
}

func (s *challengeImportTxStore) ResolvePlatformBuildImage(
	ctx context.Context,
	req challengeports.ImportedPlatformBuildImageRequest,
) (*challengeports.ImportedImageResolution, error) {
	if s.imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := s.imageBuild.CreatePlatformBuildJobInTx(
		ctx,
		&runtimeImageBuildTxStore{tx: s.tx()},
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

func (s *challengeImportTxStore) ResolveExternalImage(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	if s.imageBuild == nil {
		return nil, fmt.Errorf("image build service is not configured")
	}
	result, err := s.imageBuild.VerifyExternalImageRefInTx(
		ctx,
		&runtimeImageBuildTxStore{tx: s.tx()},
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

func (s *challengeImportTxStore) ResolveExistingImageRef(
	ctx context.Context,
	packageSlug string,
	imageRef string,
) (*challengeports.ImportedImageResolution, error) {
	ref := strings.TrimSpace(imageRef)
	if ref == "" {
		return &challengeports.ImportedImageResolution{}, nil
	}
	name, tag, err := domain.SplitImageRef(ref)
	if err != nil {
		return nil, fmt.Errorf("invalid image ref for %s: %w", packageSlug, err)
	}

	tx := s.tx().WithContext(ctx)
	var image model.Image
	findErr := tx.Unscoped().
		Where("name = ? AND tag = ?", name, tag).
		First(&image).Error
	switch {
	case errors.Is(findErr, gorm.ErrRecordNotFound):
		image = model.Image{
			Name:        name,
			Tag:         tag,
			Description: fmt.Sprintf("Imported from challenge pack %s", packageSlug),
			Status:      model.ImageStatusAvailable,
			Size:        0,
		}
		if err := tx.Create(&image).Error; err != nil {
			return nil, fmt.Errorf("create image %s:%s for %s: %w", name, tag, packageSlug, err)
		}
	case findErr != nil:
		return nil, fmt.Errorf("find image %s:%s for %s: %w", name, tag, packageSlug, findErr)
	default:
		if err := tx.Model(&image).Updates(map[string]any{
			"status":     model.ImageStatusAvailable,
			"deleted_at": nil,
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
			return nil, fmt.Errorf("update image %s:%s for %s: %w", name, tag, packageSlug, err)
		}
	}
	return &challengeports.ImportedImageResolution{
		ImageID:  image.ID,
		ImageRef: ref,
	}, nil
}
